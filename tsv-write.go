package main

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func returnToPool(b *bytes.Buffer) {
	bufPool.Put(b)
}

func WriteTSV(records <-chan *CsvRow, tsvDataOut chan<- *bytes.Buffer) {

	// TODO: Better actions if here finds errors

	var record *CsvRow

	maxRecordsInFile := 1024 * *conf.maxRecordsPerFile
	Debug.Printf("maxRecordsInFile=%v", maxRecordsInFile)

	recordsInFile := 0
	keepReading := true

	bufPointer := bufPool.Get().(*bytes.Buffer)
	bufPointer.Reset()

	gzipWriter := gzip.NewWriter(bufPointer)

	w := csv.NewWriter(gzipWriter)
	w.Comma = '\t'

	for keepReading {

		if err := w.Error(); err != nil {
			Error.Printf("TSV Writer Error: %v", err)
		}

		record, keepReading = <-records
		if keepReading {
			if recordsInFile%1000 == 0 {
				Debug.Printf("TSV Sample=%v", record)
				Debug.Printf("Records In File So Far=%v (%d / %d)", recordsInFile, bufPointer.Len(), bufPointer.Cap())
			}
			err := w.Write(*record)
			if err != nil {
				Error.Printf("TSV writer error: %+v", err)
			}

			recordsInFile++
		}
		if (recordsInFile >= maxRecordsInFile) || (!keepReading) {

			// FLUSH TSV DATA TO GZ
			w.Flush()
			if err := w.Error(); err != nil {
				Error.Printf("TSV Flush Error: %v", err)
			}
			// FLUSH GZ TO BUF
			err := gzipWriter.Flush()
			if err != nil {
				Error.Printf("GZ Flush error: %+v", err)
			}
			gzipWriter.Close()

			Info.Printf("Records In File=%v (%d)", recordsInFile, bufPointer.Len())
			if recordsInFile > 0 {
				tsvDataOut <- bufPointer
			}

			// PREPARE FOR THE NEXT BUF (Next S3 file)
			bufPointer = bufPool.Get().(*bytes.Buffer)
			bufPointer.Reset()
			gzipWriter.Reset(bufPointer)
			w = csv.NewWriter(gzipWriter)
			w.Comma = '\t'
			recordsInFile = 0
		}

	}

}
