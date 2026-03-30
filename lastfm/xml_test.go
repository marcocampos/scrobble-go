package lastfm

import (
	"io"
	"log/slog"
	"strings"
	"testing"
)

const sampleArtistXML = `<lfm status="ok">
  <artist>
    <name>Iron Maiden</name>
    <mbid>ca891d65-d9b0-4258-89f7-e6ba29d83767</mbid>
    <url>https://www.last.fm/music/Iron+Maiden</url>
    <image size="small">https://example.com/small.jpg</image>
    <image size="medium">https://example.com/medium.jpg</image>
    <image size="large">https://example.com/large.jpg</image>
    <streamable>0</streamable>
    <ontour>0</ontour>
    <stats>
      <listeners>3456789</listeners>
      <playcount>123456789</playcount>
    </stats>
    <tags>
      <tag>
        <name>heavy metal</name>
        <url>https://www.last.fm/tag/heavy+metal</url>
      </tag>
      <tag>
        <name>metal</name>
        <url>https://www.last.fm/tag/metal</url>
      </tag>
    </tags>
    <bio>
      <summary>Iron Maiden are a heavy metal band.</summary>
      <content>Long form content here.</content>
    </bio>
  </artist>
</lfm>`

const sampleErrorXML = `<lfm status="failed">
  <error code="10">Invalid API key - You must be granted a valid key by last.fm</error>
</lfm>`

func TestParseXMLResponse_OK(t *testing.T) {
	doc, err := parseXMLResponse(sampleArtistXML)
	if err != nil {
		t.Fatalf("parseXMLResponse: unexpected error: %v", err)
	}
	if doc.XMLName.Local != "lfm" {
		t.Errorf("root element = %q, want %q", doc.XMLName.Local, "lfm")
	}
	if doc.attr("status") != "ok" {
		t.Errorf("status = %q, want %q", doc.attr("status"), "ok")
	}
}

func TestParseXMLResponse_InvalidXML(t *testing.T) {
	_, err := parseXMLResponse("<not valid xml")
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}
}

func TestExtract_SimpleField(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	got := extract(doc, "name")
	if got != "Iron Maiden" {
		t.Errorf("extract(name) = %q, want %q", got, "Iron Maiden")
	}
}

func TestExtract_NestedField(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	got := extract(doc, "listeners")
	if got != "3456789" {
		t.Errorf("extract(listeners) = %q, want %q", got, "3456789")
	}
}

func TestExtract_IndexedField(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	// Second <name> tag inside <tags> should be "metal"
	names := extractAll(doc, "name")
	// names: [Iron Maiden, heavy metal, metal]
	if len(names) < 3 {
		t.Fatalf("extractAll(name) returned %d elements, want ≥3", len(names))
	}
	if names[0] != "Iron Maiden" {
		t.Errorf("names[0] = %q, want %q", names[0], "Iron Maiden")
	}
	if names[1] != "heavy metal" {
		t.Errorf("names[1] = %q, want %q", names[1], "heavy metal")
	}
	if names[2] != "metal" {
		t.Errorf("names[2] = %q, want %q", names[2], "metal")
	}
}

func TestExtract_MissingField(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	got := extract(doc, "nonexistent")
	if got != "" {
		t.Errorf("extract(nonexistent) = %q, want %q", got, "")
	}
}

func TestExtract_OutOfBoundsIndex(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	got := extract(doc, "name", 999)
	if got != "" {
		t.Errorf("extract(name, 999) = %q, want %q", got, "")
	}
}

func TestXMLNode_Attr(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	images := doc.findAll("image")
	if len(images) < 3 {
		t.Fatalf("findAll(image) = %d, want ≥3", len(images))
	}
	if images[0].attr("size") != "small" {
		t.Errorf("images[0].attr(size) = %q, want %q", images[0].attr("size"), "small")
	}
	if images[1].attr("size") != "medium" {
		t.Errorf("images[1].attr(size) = %q, want %q", images[1].attr("size"), "medium")
	}
}

func TestXMLNode_Find(t *testing.T) {
	doc, _ := parseXMLResponse(sampleArtistXML)
	stats := doc.find("stats")
	if stats == nil {
		t.Fatal("find(stats) returned nil")
	}
	listeners := extract(stats, "listeners")
	if listeners != "3456789" {
		t.Errorf("extract(listeners) from stats = %q, want %q", listeners, "3456789")
	}
}

func TestCheckAPIErrors_OK(t *testing.T) {
	err := checkAPIErrors(sampleArtistXML, "Last.fm")
	if err != nil {
		t.Errorf("checkAPIErrors on OK response: unexpected error: %v", err)
	}
}

func TestCheckAPIErrors_WSError(t *testing.T) {
	err := checkAPIErrors(sampleErrorXML, "Last.fm")
	if err == nil {
		t.Fatal("checkAPIErrors on error response: expected error, got nil")
	}
	var wsErr *WSError
	switch e := err.(type) {
	case *WSError:
		wsErr = e
	default:
		t.Fatalf("expected *WSError, got %T: %v", err, err)
	}
	if wsErr.Status != "10" {
		t.Errorf("WSError.Status = %q, want %q", wsErr.Status, "10")
	}
	if wsErr.Details == "" {
		t.Error("WSError.Details should not be empty")
	}
}

func TestCheckAPIErrors_MalformedXML(t *testing.T) {
	err := checkAPIErrors("<broken", "Last.fm")
	if err == nil {
		t.Fatal("expected error for malformed XML, got nil")
	}
	if _, ok := err.(*MalformedResponseError); !ok {
		t.Errorf("expected *MalformedResponseError, got %T", err)
	}
}

func TestCheckAPIErrors_StatusNotOkNoErrorElement(t *testing.T) {
	// status="failed" but no <error> child element.
	xml := `<lfm status="failed"></lfm>`
	err := checkAPIErrors(xml, "Last.fm")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, ok := err.(*MalformedResponseError); !ok {
		t.Errorf("expected *MalformedResponseError, got %T", err)
	}
}

func TestCheckAPIErrors_DebugLogging(t *testing.T) {
	// Install a debug-level handler so the slog.Debug branch is executed.
	h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(old)

	if err := checkAPIErrors(sampleArtistXML, "Last.fm"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCheckAPIErrors_DebugLogging_LongBody(t *testing.T) {
	// Body longer than 500 chars exercises the truncation branch inside the debug block.
	h := slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug})
	old := slog.Default()
	slog.SetDefault(slog.New(h))
	defer slog.SetDefault(old)

	body := `<lfm status="ok"><data>` + strings.Repeat("x", 600) + `</data></lfm>`
	if err := checkAPIErrors(body, "Last.fm"); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}
