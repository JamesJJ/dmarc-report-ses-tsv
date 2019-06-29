package main

import (
	//	"encoding/binary"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	//	"github.com/jhillyerd/enmime"
	"io"
	"io/ioutil"
	//"compress/gzip"
	//"archive/zip"
	"encoding/csv"
)

func main() {
	// Open a sample message file.
	r, err := ioutil.ReadFile("test-xml")
	if err != nil {
		fmt.Print(err)
		return
	}
	/*
		// Parse message body with enmime.
		env, err := enmime.ReadEnvelope(r)
		if err != nil {
			fmt.Print(err)
			return
		}

		// Headers can be retrieved via Envelope.GetHeader(name).
		//fmt.Printf("From: %v\n", env.GetHeader("From"))

		// Address-type headers can be parsed into a list of decoded mail.Address structs.
		//alist, _ := env.AddressList("To")
		//for _, addr := range alist {
		//fmt.Printf("To: %s <%s>\n", addr.Name, addr.Address)
		//}

		// enmime can decode quoted-printable headers.
		//fmt.Printf("Subject: %v\n", env.GetHeader("Subject"))

		// The plain text body is available as mime.Text.
		//fmt.Printf("Text Body: %v chars\n", len(env.Text))

		// The HTML body is stored in mime.HTML.
		//fmt.Printf("HTML Body: %v chars\n", len(env.HTML))

		// mime.Inlines is a slice of inlined attacments.
		//fmt.Printf("Inlines: %v\n", len(env.Inlines))

		// mime.Attachments contains the non-inline attachments.
		for _, attachment := range env.Attachments {
			//	  fmt.Printf("%v", attachment.Content)
			binary.Write(os.Stdout, binary.LittleEndian, attachment.Content)
		}
	*/

	xmlParse(&r)

}

func NewTSVWriter(w io.Writer) (writer *csv.Writer) {
	writer = csv.NewWriter(w)
	writer.Comma = '\t'
	return
}

/*
// Row for each IP address
type Row struct {
        SourceIP net.IP          `xml:"source_ip"`
        Count    int             `xml:"count"`
        Policy   PolicyEvaluated `xml:"policy_evaluated"`
}


// PolicyEvaluated what was evaluated
type PolicyEvaluated struct {
        Disposition string                 `xml:"disposition"`
        DKIM        string                 `xml:"dkim"`
        SPF         string                 `xml:"spf"`
        Reasons     []PolicyOverrideReason `xml:"reason,omitempty"`
}

// PolicyOverrideReason are the reasons that may affect DMARC disposition
// or execution thereof
type PolicyOverrideReason struct {
        Type    string `xml:"type",json:"type"`
        Comment string `xml:"comment",json:"comment"`
}


*/

func FlattenDMARCRecord(r *Record) [20]string {

	poReasons := ""
	poReasonsJSON, err := json.Marshal(r.Row.Policy.Reasons)
	if err == nil {
		poReasons = string(poReasonsJSON)
	}

	a := [20]string{
		r.Identifiers.HeaderFrom,
		r.Identifiers.EnvelopeFrom,
		r.Identifiers.EnvelopeFrom,

		r.AuthResults.SPF.Domain,
		r.AuthResults.SPF.Selector,
		r.AuthResults.SPF.Result,
		r.AuthResults.SPF.HumanResult,

		r.AuthResults.DKIM.Domain,
		r.AuthResults.DKIM.Selector,
		r.AuthResults.DKIM.Result,
		r.AuthResults.DKIM.HumanResult,

		r.Row.SourceIP.String(),

		strconv.Itoa(r.Row.Count),

		r.Row.Policy.Disposition,
		r.Row.Policy.DKIM,
		r.Row.Policy.SPF,
		poReasons,
	}

	return a
}

func xmlParse(body *[]byte) {

	var report Feedback

	if err := xml.Unmarshal(*body, &report); err != nil {
		fmt.Printf("%#v", err)
	}
	fmt.Printf("\n\nxml=%#v", report)
	fmt.Printf("\n\nxml=%#v", report.Records)
	for _, zz := range report.Records {
		fmt.Printf("\n\nflatten=%#v", FlattenDMARCRecord(&zz))
	}
}
