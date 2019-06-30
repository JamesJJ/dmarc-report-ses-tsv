package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	//	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/jhillyerd/enmime"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

func ReadMail(in *bytes.Reader, out chan<- *[]string) {

	Debug.Printf("Parsing email")

	// Parse email message body
	envelope, err := enmime.ReadEnvelope(in)
	if err != nil {
		fmt.Print(err)
		return
	}

	// mime.Attachments contains non-inline attachments.
	for _, attachment := range envelope.Attachments {
		Debug.Printf("%v ==> %v", attachment.ContentType, attachment.FileName)

		if attachment.ContentType == "application/zip" && strings.HasSuffix(attachment.FileName, ".zip") {
			err := UnZipBytes(&attachment.Content, out)
			if err != nil {
				Error.Printf("%v", err)
				continue
			}
		}
		if attachment.ContentType == "application/gzip" && strings.HasSuffix(attachment.FileName, ".gz") {
			err := UnGzipBytes(&attachment.Content, out)
			if err != nil {
				Error.Printf("%v", err)
				continue
			}
		}
	}

}

func UnGzipBytes(b *[]byte, out chan<- *[]string) error {
	Debug.Printf("UnGzipping Attachment")

	zr, err := gzip.NewReader(bytes.NewReader(*b))
	if err != nil {
		return fmt.Errorf("UnGzip Error: %s", err)
	}
	defer zr.Close()

	//Debug.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())
	iorc := ioutil.NopCloser(zr)

	if err := xmlParse(&iorc, out); err != nil {
		return fmt.Errorf("xmlParse Error: %s", err)
	}
	return nil
}

func UnZipBytes(b *[]byte, out chan<- *[]string) error {

	Debug.Printf("UnZipping Attachment")

	zr, err := zip.NewReader(bytes.NewReader(*b), int64(len(*b)))
	if err != nil {
		return fmt.Errorf("UnZip Error: %s", err)
	}

	for _, f := range zr.File {
		Info.Printf("Processing: %s", f.Name)
		err := xmlParseZipFile(f, out)
		if err != nil {
			Error.Printf("xmlParseZipFile Error: %v", err)
			continue
		}

	}
	return nil
}

func xmlParseZipFile(f *zip.File, out chan<- *[]string) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("xmlParseZipFile Open Error: %v", err)
	}
	defer rc.Close()
	if err := xmlParse(&rc, out); err != nil {
		return fmt.Errorf("xmlParse Error: %v (%v)", err, f.FileHeader.Name)
	}
	return nil

}

func xmlParse(body *io.ReadCloser, out chan<- *[]string) error {

	var report Feedback

	if err := xml.NewDecoder(*body).Decode(&report); err != nil {
		return fmt.Errorf("XML Decode Error: %s", err)
	}
	for rrIndex, rrRecord := range report.Records {
		fdr, err := FlattenDMARCRecord(&report.Metadata, &report.Policy, &rrRecord)
		if err != nil {
			Error.Printf("FlattenDMARCRecord Error: %v", err)
			continue
		}
		out <- &fdr
		if rrIndex%1000 == 0 {
			Debug.Printf("Flattened DMARC Sample=%v", fdr)
		}
	}

	return nil

}

func timestampToString(t int64) (s string) {
	theTime := time.Unix(t, 0)
	return strings.Replace(theTime.UTC().Format(time.RFC3339), "T", " ", 1)[0:19]
}

func FlattenDMARCRecord(m *ReportMetadata, pp *PolicyPublished, r *Record) ([]string, error) {

	var poReasons strings.Builder
	var poReasonsComments strings.Builder
	for _, por := range r.Row.Policy.Reasons {
		if por.Type != "" {
			if poReasons.Len() > 0 {
				poReasons.WriteString("|")
			}
			poReasons.WriteString(por.Type)
		}
		if por.Comment != "" {
			if poReasonsComments.Len() > 0 {
				poReasonsComments.WriteString("|")
			}
			poReasonsComments.WriteString(por.Comment)
		}
	}

	var mErrorString strings.Builder
	for _, mError := range m.Errors {
		if mErrorString.Len() > 0 {
			mErrorString.WriteString("|")
		}
		mErrorString.WriteString(mError)

	}

	a := []string{

		m.ReportID,
		timestampToString(m.Date.Begin),
		timestampToString(m.Date.End),
		m.OrgName,
		m.Email,
		m.ExtraContactInfo,
		mErrorString.String(),

		pp.Domain,
		pp.ASPF,
		pp.ADKIM,
		pp.P,
		pp.SP,
		strconv.Itoa(pp.Pct),
		pp.Fo,

		r.Identifiers.HeaderFrom,
		r.Identifiers.EnvelopeFrom,
		r.Identifiers.EnvelopeTo,

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
		poReasons.String(),
		poReasonsComments.String(),
	}

	for i := range a {
		a[i] = strings.ReplaceAll(a[i], "\t", "")
		a[i] = strings.ReplaceAll(a[i], "\r", "")
		a[i] = strings.ReplaceAll(a[i], "\n", "")
	}

	return a, nil
}
