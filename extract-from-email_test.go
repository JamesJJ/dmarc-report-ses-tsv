package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

type FileTest struct {
	File         string
	ExpectedRows int
	Rows         CsvRows
}

func loadTestData(t *testing.T, filename string) *bytes.Reader {

	var err error

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	fileinfo, err := file.Stat()
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	filesize := fileinfo.Size()
	buffer := make([]byte, filesize)
	_, err = file.Read(buffer)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		os.Exit(1)
	}

	return bytes.NewReader(buffer)
}

func TestReadMailToRows(t *testing.T) {

	fileTests := []FileTest{
		{
			File:         "./test-data/mail-with-bad-gzip.txt",
			ExpectedRows: 0,
			Rows:         CsvRows{},
		}, {
			File:         "./test-data/mail-with-bad-zip.txt",
			ExpectedRows: 0,
			Rows:         CsvRows{},
		}, {
			File:         "./test-data/mail-with-non-dmarc-zip.txt",
			ExpectedRows: 0,
			Rows:         CsvRows{},
		}, {
			File:         "./test-data/mail-without-attachments.txt",
			ExpectedRows: 0,
			Rows:         CsvRows{},
		},
		{
			File:         "./test-data/mail-with-zipped-dmarc-report.txt",
			ExpectedRows: 1,
			Rows: CsvRows{
				&CsvRow{
					"dmarc@example.com",
					"1442744299377440478",
					"2021-03-05 00:00:00",
					"2021-03-05 23:59:59",
					"google.com",
					"noreply-dmarc-support@google.com",
					"https://support.google.com/a/answer/2466580",
					"",
					"example.com",
					"r",
					"r",
					"reject",
					"reject",
					"100",
					"",
					"example.com",
					"",
					"",
					"a27-5.smtp-out.us-west-2.amazonses.com",
					"",
					"none",
					"",
					"amazonses.com",
					"hsbnp8p3ensaochzwyq5wwmceodymuwv",
					"pass",
					"",
					"54.240.27.5",
					"1",
					"reject",
					"fail",
					"fail",
					"",
					"",
				},
			},
		},
		{
			File:         "./test-data/mail-with-gzip-dmarc-report.txt",
			ExpectedRows: 1,
			Rows: CsvRows{
				&CsvRow{
					"dmarc@example.com",
					"1442744299377440477",
					"2021-03-05 00:00:00",
					"2021-03-05 23:59:59",
					"google.com",
					"noreply-dmarc-support@google.com",
					"https://support.google.com/a/answer/2466580",
					"",
					"example.com",
					"r",
					"r",
					"reject",
					"reject",
					"100",
					"",
					"example.com",
					"",
					"",
					"a27-5.smtp-out.us-west-2.amazonses.com",
					"",
					"none",
					"",
					"amazonses.com",
					"hsbnp8p3ensaochzwyq5wwmceodymuwv",
					"pass",
					"",
					"54.240.27.5",
					"1",
					"reject",
					"fail",
					"fail",
					"",
					"",
				},
			},
		},
	}

	for _, ft := range fileTests {

		t.Run(fmt.Sprintf("Input data: %s", ft.File), func(t *testing.T) {

			in := loadTestData(t, ft.File)
			csvRows := *ReadMailToRows(in)

			rowCount := len(csvRows)
			if rowCount != ft.ExpectedRows {
				t.Errorf("Input %s: Expected %d rows, got %d", ft.File, ft.ExpectedRows, rowCount)
			}

			for rowIndex, desiredRow := range ft.Rows {
				gotColumns := *csvRows[rowIndex]
				desiredColumns := *ft.Rows[rowIndex]
				for columnIndex := range *desiredRow {
					if gotColumns[columnIndex] != desiredColumns[columnIndex] {
						t.Errorf(
							"Input %s: Row %d, Column %d. Expected %s, got %s",
							ft.File,
							rowIndex,
							columnIndex,
							desiredColumns[columnIndex],
							gotColumns[columnIndex],
						)
					}
				}
			}
		})
	}

}
