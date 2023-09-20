package cmd

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/fatih/color"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	hostUrl      = "https://scsservices2.azurewebsites.net/SCSService.asmx"
	xmlLoginBody = `
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <SystemObjectRun xmlns="http://scsservices.net/">
      <args>
        <SessionGuid>00000000-0000-0000-0000-000000000000</SessionGuid>
        <Action>Login</Action>
        <Format>Xml</Format>
        <Bytes/>
        <Value>&lt;?xml version="1.0" encoding="utf-16"?&gt;
&lt;TLoginInputArgs xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"&gt;
  &lt;AppName&gt;App&lt;/AppName&gt;
  &lt;CompanyID&gt;SCS076&lt;/CompanyID&gt;
  &lt;UserID&gt;{employee}&lt;/UserID&gt;
  &lt;Password&gt;{identity}&lt;/Password&gt;
  &lt;LanguageID&gt;zh-TW&lt;/LanguageID&gt;
  &lt;UserHostAddress /&gt;
  &lt;IsSaveSessionBuffer&gt;true&lt;/IsSaveSessionBuffer&gt;
  &lt;ValidateCode /&gt;
  &lt;OAuthType&gt;NotSet&lt;/OAuthType&gt;
  &lt;IsValidateRegister&gt;false&lt;/IsValidateRegister&gt;
&lt;/TLoginInputArgs&gt;</Value>
      </args>
    </SystemObjectRun>
  </soap12:Body>
</soap12:Envelope>`

	xmlClockBody = `
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <BusinessObjectRun xmlns="http://scsservices.net/">
      <args>
        <SessionGuid>{sessionGuid}</SessionGuid>
        <ProgID>WATT0022000</ProgID>
        <Action>ExecFunc</Action>
        <Format>Xml</Format>
        <Bytes/>
        <Value>&lt;?xml version="1.0" encoding="utf-16"?&gt;
&lt;TExecFuncInputArgs xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"&gt;
  &lt;FuncID&gt;ExecuteSwipeData_Web&lt;/FuncID&gt;
  &lt;Parameters&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;DutyCode&lt;/Name&gt;
      &lt;Value xsi:type="xsd:int"&gt;{dutyCode}&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;DutyStatus&lt;/Name&gt;
      &lt;Value xsi:type="xsd:int"&gt;{dutyStatus}&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;GPSLocation&lt;/Name&gt;
      &lt;Value xsi:type="xsd:string"&gt;{latitude},{longitude}&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;CompanyID&lt;/Name&gt;
      &lt;Value xsi:type="xsd:string"&gt;SCS076&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;GpsAddress&lt;/Name&gt;
      &lt;Value xsi:type="xsd:string"&gt;{address}&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;NoCheckOnDutyStatus&lt;/Name&gt;
      &lt;Value xsi:type="xsd:boolean"&gt;true&lt;/Value&gt;
    &lt;/Parameter&gt;
  &lt;/Parameters&gt;
&lt;/TExecFuncInputArgs&gt;</Value>
      </args>
    </BusinessObjectRun>
  </soap12:Body>
</soap12:Envelope>`

	xmlGetRecordBody = `
<soap12:Envelope xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <BusinessObjectRun xmlns="http://scsservices.net/">
      <args>
        <SessionGuid>{sessionGuid}</SessionGuid>
        <ProgID>EIP0080600</ProgID>
        <Action>ExecFunc</Action>
        <Format>Xml</Format>
        <Bytes/>
        <Value>&lt;?xml version="1.0" encoding="utf-16"?&gt;
&lt;TExecFuncInputArgs xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"&gt;
  &lt;FuncID&gt;GetSwipeQueryListData&lt;/FuncID&gt;
  &lt;Parameters&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;setting&lt;/Name&gt;
      &lt;Value xsi:type="xsd:string"&gt;{
  "WebPartID": "SwipeQuery",
  "DisplayName": "個人刷卡資料查詢",
  "FuncType": 2,
  "FuncID": "GetSwipeQueryListData",
  "IsEnabledGrouping": true,
  "Number": ""
}&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;caller&lt;/Name&gt;
      &lt;Value xsi:type="xsd:string"&gt;SCSAppForms&lt;/Value&gt;
    &lt;/Parameter&gt;
    &lt;Parameter&gt;
      &lt;Name&gt;SelectCount&lt;/Name&gt;
      &lt;Value xsi:type="xsd:int"&gt;20&lt;/Value&gt;
    &lt;/Parameter&gt;
  &lt;/Parameters&gt;
&lt;/TExecFuncInputArgs&gt;</Value>
      </args>
    </BusinessObjectRun>
  </soap12:Body>
</soap12:Envelope>
`

	clockInDutyCode    = "4"
	clockInDutyStatus  = "4"
	clockOffDutyCode   = "8"
	clockOffDutyStatus = "5"
)

type loginResult struct {
	XMLName     xml.Name `xml:"TLoginOutputResult"`
	SessionGuid string   `xml:"SessionGuid"`
	SessionInfo interface{}
}

type recordParameter struct {
	XMLName xml.Name `xml:"Parameter"`
	Value   []byte   `xml:"Value"`
}

type clockInOffData struct {
	Data []record `json:"Data"`
}

type record struct {
	AttendDate  string `json:"ATTENDDATE"`
	DisplayName string `json:"SWIPEDATEDISPLAYNAME"`
}

type recordParameters struct {
	XMLName   xml.Name        `xml:"Parameters"`
	Parameter recordParameter `xml:"Parameter"`
}

type recordResult struct {
	XMLName    xml.Name         `xml:"TExecFuncOutputResult"`
	Parameters recordParameters `xml:"Parameters"`
}

type result struct {
	XMLName xml.Name `xml:"SystemObjectRunResult"`
	Action  string   `xml:"Action"`
	Format  string   `xml:"XML"`
	Value   []byte   `xml:"Value"`
}

type businessResult struct {
	XMLName xml.Name `xml:"BusinessObjectRunResult"`
	Value   []byte   `xml:"Value"`
}

type response struct {
	XMLName xml.Name `xml:"SystemObjectRunResponse"`
	Result  result   `xml:"SystemObjectRunResult"`
}

type businessResponse struct {
	XMLName xml.Name       `xml:"BusinessObjectRunResponse"`
	Result  businessResult `xml:"BusinessObjectRunResult"`
}

type body struct {
	XMLName  xml.Name `xml:"Body"`
	Response response `xml:"SystemObjectRunResponse"`
}

type businessBody struct {
	XMLName  xml.Name         `xml:"Body"`
	Response businessResponse `xml:"BusinessObjectRunResponse"`
}

type envelope struct {
	XMLName  xml.Name `xml:"Envelope"`
	SoapBody body
}

type businessEnvelope struct {
	XMLName  xml.Name `xml:"Envelope"`
	SoapBody businessBody
}

func login(employee, identity string) string {
	body := strings.Replace(xmlLoginBody, "{employee}", employee, 1)
	body = strings.Replace(body, "{identity}", identity, 1)
	req, _ := http.NewRequest("POST", hostUrl, strings.NewReader(body))
	req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Length", "1123")
	req.Header.Add("HOST", "scsservices2.azurewebsites.net")
	res, _ := http.DefaultClient.Do(req)
	content, _ := ioutil.ReadAll(res.Body)
	v := envelope{}
	err := xml.Unmarshal(content, &v)
	if err != nil {
		fmt.Printf("error: %v", err)
		return ""
	}
	value := v.SoapBody.Response.Result.Value

	value = bytes.Replace(value, []byte("utf-16"), []byte("utf-8"), 1)

	loginResult := loginResult{}
	if err := xml.Unmarshal(value, &loginResult); err != nil {
		fmt.Printf("error: %v", err)
		return ""
	}

	return loginResult.SessionGuid
}

func clockIn(sessionID, latitude, longitude, address string) {
	url := strings.Replace(xmlClockBody, "{sessionGuid}", sessionID, 1)
	url = strings.Replace(url, "{dutyStatus}", clockInDutyStatus, 1)
	url = strings.Replace(url, "{dutyCode}", clockInDutyCode, 1)
	url = strings.Replace(url, "{latitude}", latitude, 1)
	url = strings.Replace(url, "{longitude}", longitude, 1)
	url = strings.Replace(url, "{address}", address, 1)
	req, _ := http.NewRequest("POST", hostUrl, strings.NewReader(url))
	req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Length", "1123")
	req.Header.Add("HOST", "scsservices2.azurewebsites.net")
	_, err := http.DefaultClient.Do(req)
	if err != nil {
	} else {
		color.Green("Clock in success")
		getTodayRecord(sessionID)

	}
}

func clockOff(sessionID, latitude, longitude, address string) {
	url := strings.Replace(xmlClockBody, "{sessionGuid}", sessionID, 1)
	url = strings.Replace(url, "{dutyStatus}", clockOffDutyStatus, 1)
	url = strings.Replace(url, "{dutyCode}", clockOffDutyCode, 1)
	url = strings.Replace(url, "{latitude}", latitude, 1)
	url = strings.Replace(url, "{longitude}", longitude, 1)
	url = strings.Replace(url, "{address}", address, 1)
	req, _ := http.NewRequest("POST", hostUrl, strings.NewReader(url))
	req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Length", "1123")
	req.Header.Add("HOST", "scsservices2.azurewebsites.net")
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		color.Red("Clock off failed")
	} else {
		color.Green("Clock off success")
		getTodayRecord(sessionID)
	}
}

func getTodayRecord(sessionID string) {
	url := strings.Replace(xmlGetRecordBody, "{sessionGuid}", sessionID, 1)
	req, _ := http.NewRequest("POST", hostUrl, strings.NewReader(url))
	req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Length", "1123")
	req.Header.Add("HOST", "scsservices2.azurewebsites.net")

	res, _ := http.DefaultClient.Do(req)

	content, _ := ioutil.ReadAll(res.Body)
	v := businessEnvelope{}
	err := xml.Unmarshal(content, &v)
	if err != nil {
		log.Fatal(err)
		return
	}
	value := v.SoapBody.Response.Result.Value

	value = bytes.Replace(value, []byte("utf-16"), []byte("utf-8"), 1)

	recordResult := recordResult{}
	if err := xml.Unmarshal(value, &recordResult); err != nil {
		log.Fatal(err.Error())
		return
	}

	clockData := recordResult.Parameters.Parameter.Value
	clockInOffData := clockInOffData{}
	if err := json.Unmarshal(clockData, &clockInOffData); err != nil {
		log.Fatal(err.Error())
		return
	}

	now := time.Now()
	nowStr := now.Format("2006/01/02")

	color.Yellow("Today's record:")
	haveRecord := false
	for _, data := range clockInOffData.Data {
		if data.AttendDate == nowStr {
			fmt.Println("----------------")
			fmt.Println(data.DisplayName)
			haveRecord = true
		}
	}
	if !haveRecord {
		fmt.Println("----------------")
		color.Red("No record")
	}
	fmt.Println("----------------")
}
