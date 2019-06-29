/*

This originally from:
https://github.com/keltia/dmarc-cat/blob/0690798d36ec52f9a7dfa0a8c3390829f831524f/types.go

BSD 2-Clause License:

Copyright (c) 2018, Ollivier Robert All rights reserved.

Redistribution and use in source and binary forms, with or without modification, are permitted
provided that the following conditions are met:

Redistributions of source code must retain the above copyright notice, this list of conditions
and the following disclaimer.

Redistributions in binary form must reproduce the above copyright notice, this list of
conditions and the following disclaimer in the documentation and/or other materials provided
with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS
OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY
AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER
OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE
OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.

The views and conclusions contained in the software and documentation are those of the
authors and should not be interpreted as representing official policies, either
expressed or implied, of the author.

*/

package main

import (
	"net"
)

// DateRange time period
type DateRange struct {
	Begin int64 `xml:"begin"`
	End   int64 `xml:"end"`
}

// ReportMetadata for the report
type ReportMetadata struct {
	OrgName          string    `xml:"org_name"`
	Email            string    `xml:"email"`
	ExtraContactInfo string    `xml:"extra_contact_info"`
	ReportID         string    `xml:"report_id"`
	Date             DateRange `xml:"date_range"`
	Errors           []string  `xml:"error"`
}

// PolicyPublished found in DNS
type PolicyPublished struct {
	Domain string `xml:"domain"`
	ADKIM  string `xml:"adkim"`
	ASPF   string `xml:"aspf"`
	P      string `xml:"p"`
	SP     string `xml:"sp"`
	Pct    int    `xml:"pct"`
	Fo     string `xml:"fo"`
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
	Type    string `xml:"type" json:"type"`
	Comment string `xml:"comment" json:"comment"`
}

// Row for each IP address
type Row struct {
	SourceIP net.IP          `xml:"source_ip"`
	Count    int             `xml:"count"`
	Policy   PolicyEvaluated `xml:"policy_evaluated"`
}

// Identifiers headers checked
type Identifiers struct {
	HeaderFrom   string `xml:"header_from"`
	EnvelopeFrom string `xml:"envelope_from"`
	EnvelopeTo   string `xml:"envelope_to,omitempty"`
}

// Result for each IP
type Result struct {
	Domain      string `xml:"domain"`
	Selector    string `xml:"selector"`
	Result      string `xml:"result"`
	HumanResult string `xml:"human_result"`
}

// AuthResults for DKIM/SPF
type AuthResults struct {
	DKIM Result `xml:"dkim,omitempty"`
	SPF  Result `xml:"spf,omitempty"`
}

// Record for each IP
type Record struct {
	Row         Row         `xml:"row"`
	Identifiers Identifiers `xml:"identifiers"`
	AuthResults AuthResults `xml:"auth_results"`
}

// Feedback the report itself
type Feedback struct {
	Version  float32         `xml:"version"`
	Metadata ReportMetadata  `xml:"report_metadata"`
	Policy   PolicyPublished `xml:"policy_published"`
	Records  []Record        `xml:"record"`
}
