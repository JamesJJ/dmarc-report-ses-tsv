package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"github.com/jhillyerd/enmime"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

type AttachmentCompression int

const (
	NotCompressed AttachmentCompression = iota
	Zipped
	Gzipped
)

type LabeledMailPart struct {
	Compression AttachmentCompression
	Part        *enmime.Part
}

type MailParts struct {
	MailToFirstAddress string
	AttachedParts      []LabeledMailPart
}

type DecompressedReader struct {
	ReadCloser *io.ReadCloser
	CloseFunc  func()
}

type CsvRow []string

type CsvRows []*CsvRow

func ReadMail(in *bytes.Reader, out chan<- *CsvRow) {
	for _, rowPointer := range *ReadMailToRows(in) {
		out <- rowPointer
	}
}

func ReadMailToRows(in *bytes.Reader) *CsvRows {

	var csvRows CsvRows

	mailParts, err := ReadRawMail(in)
	if err != nil {
		Error.Printf("%v", err)
		return &csvRows
	}

	decompressedReaders := getDecompressedReaders(mailParts)
	for _, reader := range decompressedReaders {

		report, xmlErr := dmarcXmlParse(reader.ReadCloser)
		reader.CloseFunc()
		if xmlErr != nil {
			Error.Printf("%v", xmlErr)
			continue
		}
		csvRows = append(csvRows, *flattenDMARCRecord(mailParts.MailToFirstAddress, report)...)
	}
	return &csvRows
}

func ReadRawMail(in *bytes.Reader) (*MailParts, error) {

	Debug.Printf("Parsing email")

	// Parse email message body
	envelope, err := enmime.ReadEnvelope(in)
	if err != nil {
		fmt.Print(err)
		return nil, err
	}

	msgToFirstAddress := "unknown"
	if alist, alistErr := envelope.AddressList("To"); alistErr == nil {
		for _, addr := range alist {
			msgToFirstAddress = addr.Address
			break
		}
	}

	mail := MailParts{
		MailToFirstAddress: msgToFirstAddress,
	}

	// mime.Attachments contains non-inline attachments.
	for _, attachment := range envelope.Attachments {
		Debug.Printf("%v ==> %v", attachment.ContentType, attachment.FileName)

		var compression AttachmentCompression

		if attachment.ContentType == "application/zip" && strings.HasSuffix(attachment.FileName, ".zip") {
			compression = Zipped
		} else if attachment.ContentType == "application/gzip" && strings.HasSuffix(attachment.FileName, ".gz") {
			compression = Gzipped
		} else {
			continue
		}

		mail.AttachedParts = append(
			mail.AttachedParts,
			LabeledMailPart{
				Compression: compression,
				Part:        attachment,
			},
		)
	}
	return &mail, nil
}

func getDecompressedReaders(mailparts *MailParts) []*DecompressedReader {
	var decompressedReaders []*DecompressedReader
	for _, part := range mailparts.AttachedParts {
		Debug.Printf("%v ==> %v", part.Part.ContentType, part.Part.FileName)

		if part.Compression == Zipped {
			if readers, err := zipFileReaders(&part.Part.Content); err == nil {
				decompressedReaders = append(
					decompressedReaders,
					readers...,
				)
			} else {
				Error.Printf("%v", err)
				continue
			}
		}

		if part.Compression == Gzipped {
			Info.Printf("Processing Gzipped: %s", part.Part.FileName)
			if reader, err := gzipReader(&part.Part.Content); err == nil {
				decompressedReaders = append(
					decompressedReaders,
					reader,
				)
			} else {
				Error.Printf("%v", err)
				continue
			}
		}
	}
	return decompressedReaders
}

func gzipReader(b *[]byte) (*DecompressedReader, error) {
	Debug.Printf("UnGzipping Attachment")

	zr, err := gzip.NewReader(bytes.NewReader(*b))
	if err != nil {
		return nil, fmt.Errorf("UnGzip Error: %s", err)
	}

	//Debug.Printf("Name: %s\nComment: %s\nModTime: %s\n\n", zr.Name, zr.Comment, zr.ModTime.UTC())
	iorc := ioutil.NopCloser(zr)
	return &DecompressedReader{
		ReadCloser: &iorc,
		CloseFunc:  func() { zr.Close() },
	}, nil
}

func zipFileReaders(b *[]byte) ([]*DecompressedReader, error) {

	Debug.Printf("UnZipping Attachment")

	var readers []*DecompressedReader

	zr, err := zip.NewReader(bytes.NewReader(*b), int64(len(*b)))
	if err != nil {
		return readers, fmt.Errorf("UnZip Error: %s", err)
	}

	for _, f := range zr.File {
		Info.Printf("Processing UnZipped: %s", f.Name)
		if rc, err := readZipEachFile(f); err == nil {
			readers = append(readers, rc)
		} else {
			Error.Printf("xmlParseZipFile Error: %v", err)
			continue
		}
	}
	return readers, nil
}

func readZipEachFile(f *zip.File) (*DecompressedReader, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, fmt.Errorf("xmlParseZipFile Open Error: %v", err)
	}
	return &DecompressedReader{
		ReadCloser: &rc,
		CloseFunc:  func() { rc.Close() },
	}, nil
}

func dmarcXmlParse(body *io.ReadCloser) (*Feedback, error) {
	var report Feedback
	if err := xml.NewDecoder(*body).Decode(&report); err != nil {
		return nil, fmt.Errorf("XML Decode Error: %s", err)
	}
	return &report, nil
}

func timestampToString(t int64) (s string) {
	theTime := time.Unix(t, 0)
	return strings.Replace(theTime.UTC().Format(time.RFC3339), "T", " ", 1)[0:19]
}

func DelimitedAppend(delimiter string, sb *strings.Builder, s string) {
	if s == "" {
		return
	}
	if sb.Len() > 0 {
		sb.WriteString(delimiter)
	}
	sb.WriteString(s)
}

func flattenDMARCRecord(msgToFirstAddress string, report *Feedback) *CsvRows {
	var rows CsvRows
	for rrIndex, rrRecord := range report.Records {
		if *conf.excludeDispositionNone && strings.EqualFold(rrRecord.Row.Policy.Disposition, "none") {
			continue
		}
		singleRow, err := generateDMARCRow(msgToFirstAddress, &report.Metadata, &report.Policy, &rrRecord)
		if err != nil {
			Error.Printf("flattenDMARCRecord Error: %v", err)
			continue
		}
		rows = append(rows, singleRow)
		if rrIndex%1000 == 0 {
			Debug.Printf("Flattened DMARC Sample=%v", singleRow)
		}
	}
	return &rows
}

func generateDMARCRow(msgToFirstAddress string, m *ReportMetadata, pp *PolicyPublished, r *Record) (*CsvRow, error) {

	var poReasons strings.Builder
	var poReasonsComments strings.Builder
	for _, por := range r.Row.Policy.Reasons {
		DelimitedAppend("|", &poReasons, por.Type)
		DelimitedAppend("|", &poReasonsComments, por.Comment)
	}

	var mErrorString strings.Builder
	for _, mError := range m.Errors {
		DelimitedAppend("|", &mErrorString, mError)
	}

	// TODO: sender hits e.g.
	// r.AuthResults.SPF.Domain =>
	// https://assets.ctfassets.net/yzco4xsimv0y/2iI1A72XqE0McQcEyKaGCU/f2daf2f06c3f18a9859c4ecb5fb17f42/sending_domains.txt

	row := CsvRow{
		msgToFirstAddress,
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

	for i := range row {
		row[i] = strings.ReplaceAll(row[i], "\t", "")
		row[i] = strings.ReplaceAll(row[i], "\r", "")
		row[i] = strings.ReplaceAll(row[i], "\n", "")
	}

	return &row, nil
}
