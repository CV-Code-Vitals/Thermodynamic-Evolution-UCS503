package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"
	"unicode/utf16"
)

func TestValidateUploadAcceptsSupportedFiles(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		signature   []byte
	}{
		{
			name:        "report.pdf",
			contentType: "application/pdf",
			signature:   []byte("%PDF-1.7\n"),
		},
		{
			name:        "slides.ppt",
			contentType: "application/vnd.ms-powerpoint",
			signature:   minimalPPT(t),
		},
		{
			name:        "slides.pptx",
			contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
			signature:   minimalPPTX(t),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := validateUpload(test.name, test.contentType, int64(len(test.signature)), test.signature); err != nil {
				t.Fatalf("validateUpload rejected supported file: %v", err)
			}
		})
	}
}

func TestValidateUploadRejectsGenericOLEFileMasqueradingAsPowerPoint(t *testing.T) {
	content := minimalCompoundFile("WordDocument")
	if err := validateUpload("report.ppt", "application/vnd.ms-powerpoint", int64(len(content)), content); err == nil {
		t.Fatal("validateUpload accepted a generic OLE document as PowerPoint")
	}
}

func TestValidateUploadRejectsUnsupportedOrMismatchedFiles(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		size        int64
		signature   []byte
	}{
		{name: "file.pdf", contentType: "application/zip", size: 8, signature: []byte("%PDF-1.7")},
		{name: "file.txt", contentType: "application/pdf", size: 8, signature: []byte("%PDF-1.7")},
		{name: "file.pdf", contentType: "application/pdf", size: 8, signature: []byte("not a PDF")},
		{name: "slides.pptx", contentType: "application/vnd.openxmlformats-officedocument.presentationml.presentation", size: 8, signature: arbitraryZIP(t)},
		{name: "../report.pdf", contentType: "application/pdf", size: 8, signature: []byte("%PDF-1.7")},
		{name: `..\report.pdf`, contentType: "application/pdf", size: 8, signature: []byte("%PDF-1.7")},
		{name: "empty.pdf", contentType: "application/pdf", size: 0, signature: []byte("%PDF-1.7")},
		{name: "large.pdf", contentType: "application/pdf", size: 52_428_801, signature: []byte("%PDF-1.7")},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s_%s", test.name, test.contentType), func(t *testing.T) {
			if err := validateUpload(test.name, test.contentType, test.size, test.signature); err == nil {
				t.Fatal("validateUpload accepted invalid file metadata or content")
			}
		})
	}
}

func TestValidateArchiveMetadataEnforcesUTF8ByteLimits(t *testing.T) {
	for _, test := range []struct {
		field string
		value string
	}{
		{field: "title", value: string(bytes.Repeat([]byte("x"), 201))},
		{field: "version", value: string(bytes.Repeat([]byte("x"), 65))},
		{field: "summary", value: string(bytes.Repeat([]byte("x"), 2001))},
	} {
		t.Run(test.field, func(t *testing.T) {
			title, version, summary := "Report", "1.0", "Summary"
			switch test.field {
			case "title":
				title = test.value
			case "version":
				version = test.value
			case "summary":
				summary = test.value
			}
			if err := validateArchiveMetadata(title, version, "2026-10-22", summary); err == nil {
				t.Fatalf("validateArchiveMetadata accepted a %s over its byte limit", test.field)
			}
		})
	}
}

func TestValidateArchiveMetadataRejectsInvalidISODate(t *testing.T) {
	for _, date := range []string{"", "2026-02-30", "10/22/2026", "2026-1-2"} {
		t.Run(date, func(t *testing.T) {
			if err := validateArchiveMetadata("Report", "1.0", date, "Summary"); err == nil {
				t.Fatalf("validateArchiveMetadata accepted invalid date %q", date)
			}
		})
	}
}

func minimalPPTX(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	for _, entry := range []struct {
		name    string
		content string
	}{
		{
			name:    "[Content_Types].xml",
			content: `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/></Types>`,
		},
		{name: "ppt/presentation.xml", content: `<p:presentation xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"></p:presentation>`},
	} {
		file, err := writer.Create(entry.name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := file.Write([]byte(entry.content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func minimalPPT(t *testing.T) []byte {
	t.Helper()
	return minimalCompoundFile("PowerPoint Document")
}

func minimalCompoundFile(streamName string) []byte {
	const sectorSize = 512
	const endOfChain = uint32(0xfffffffe)
	const fatSector = uint32(0xfffffffd)
	data := make([]byte, sectorSize*4)
	copy(data[:8], []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1})
	binary.LittleEndian.PutUint16(data[24:], 0x003e)
	binary.LittleEndian.PutUint16(data[26:], 3)
	binary.LittleEndian.PutUint16(data[28:], 0xfffe)
	binary.LittleEndian.PutUint16(data[30:], 9)
	binary.LittleEndian.PutUint16(data[32:], 6)
	binary.LittleEndian.PutUint32(data[44:], 1)
	binary.LittleEndian.PutUint32(data[48:], 0)
	binary.LittleEndian.PutUint32(data[56:], 4096)
	binary.LittleEndian.PutUint32(data[60:], endOfChain)
	binary.LittleEndian.PutUint32(data[68:], endOfChain)
	binary.LittleEndian.PutUint32(data[76:], 1)
	for i := 1; i < 109; i++ {
		binary.LittleEndian.PutUint32(data[76+i*4:], 0xffffffff)
	}

	directory := data[sectorSize : sectorSize*2]
	writeCompoundDirectoryEntry(directory[:128], "Root Entry", 5, 1, 0, 0)
	writeCompoundDirectoryEntry(directory[128:256], streamName, 2, 0xffffffff, 2, 8)
	for i := 0; i < sectorSize/4; i++ {
		binary.LittleEndian.PutUint32(data[sectorSize*2+i*4:], 0xffffffff)
	}
	binary.LittleEndian.PutUint32(data[sectorSize*2:], endOfChain)
	binary.LittleEndian.PutUint32(data[sectorSize*2+4:], fatSector)
	binary.LittleEndian.PutUint32(data[sectorSize*2+8:], endOfChain)
	copy(data[sectorSize*3-8:], []byte("PPTDATA!"))
	return data
}

func writeCompoundDirectoryEntry(entry []byte, name string, kind byte, child, firstSector uint32, size uint64) {
	encodedName := utf16.Encode([]rune(name + "\x00"))
	for i, unit := range encodedName {
		binary.LittleEndian.PutUint16(entry[i*2:], unit)
	}
	binary.LittleEndian.PutUint16(entry[64:], uint16(len(encodedName)*2))
	entry[66] = kind
	entry[67] = 1
	binary.LittleEndian.PutUint32(entry[68:], 0xffffffff)
	binary.LittleEndian.PutUint32(entry[72:], 0xffffffff)
	binary.LittleEndian.PutUint32(entry[76:], child)
	binary.LittleEndian.PutUint32(entry[116:], firstSector)
	binary.LittleEndian.PutUint64(entry[120:], size)
}

func arbitraryZIP(t *testing.T) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	file, err := writer.Create("arbitrary.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("not a PowerPoint document")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}
