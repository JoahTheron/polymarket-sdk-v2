package gamma

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGammaIntegerIDEndpoints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markets/123":
			_, _ = w.Write([]byte(`{"id":123,"conditionId":"0xcond","question":"Question?"}`))
		case "/events/234":
			_, _ = w.Write([]byte(`{"id":"234","ticker":"evt"}`))
		case "/series/345":
			_, _ = w.Write([]byte(`{"id":345,"ticker":"series"}`))
		case "/tags/456":
			_, _ = w.Write([]byte(`{"id":"456","label":"tag"}`))
		case "/tags/456/related-tags":
			if r.URL.Query().Get("status") != "active" {
				t.Fatalf("related-tag query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[{"id":1,"tagID":"456","relatedTagID":789,"rank":"2"}]`))
		case "/tags/456/related-tags/tags":
			_, _ = w.Write([]byte(`[{"id":789,"label":"related"}]`))
		case "/comments/567":
			_, _ = w.Write([]byte(`[{"id":"567","eventId":234}]`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	client := New(Config{Host: srv.URL})
	market := Market{ID: 123}
	if err := client.GetMarket(context.Background(), &market); err != nil || int(market.ID) != 123 {
		t.Fatalf("market=%+v err=%v", market, err)
	}
	event := Event{ID: 234}
	if err := client.GetEvent(context.Background(), &event); err != nil || int(event.ID) != 234 {
		t.Fatalf("event=%+v err=%v", event, err)
	}
	series := Series{ID: 345}
	if err := client.GetSeries(context.Background(), &series); err != nil || int(series.ID) != 345 {
		t.Fatalf("series=%+v err=%v", series, err)
	}
	tag := Tag{ID: 456}
	if err := client.GetTag(context.Background(), &tag); err != nil || int(tag.ID) != 456 {
		t.Fatalf("tag=%+v err=%v", tag, err)
	}
	relationships, err := client.GetRelatedTagRelationships(context.Background(), 456, RelatedTagParams{Status: "active"})
	if err != nil {
		t.Fatal(err)
	}
	if len(relationships) != 1 || int(relationships[0].TagID) != 456 || int(relationships[0].Rank) != 2 {
		t.Fatalf("relationships=%+v", relationships)
	}
	related, err := client.GetRelatedTags(context.Background(), 456, RelatedTagParams{})
	if err != nil {
		t.Fatal(err)
	}
	if len(related) != 1 || int(related[0].ID) != 789 {
		t.Fatalf("related=%+v", related)
	}
	comments, err := client.GetComment(context.Background(), 567)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 || int(comments[0].ID) != 567 || int(comments[0].EventID) != 234 {
		t.Fatalf("comments=%+v", comments)
	}
}

func TestGammaIntegerIDFilters(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/markets":
			if r.URL.Query().Get("tag_id") != "12" || r.URL.Query().Get("event_id") != "34" {
				t.Fatalf("market query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[]`))
		case "/series":
			if r.URL.Query().Get("tag_id") != "56" {
				t.Fatalf("series query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[]`))
		case "/comments":
			if r.URL.Query().Get("event_id") != "78" {
				t.Fatalf("comments query = %s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`[]`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.String())
		}
	}))
	defer srv.Close()

	client := New(Config{Host: srv.URL})
	if _, err := client.GetMarkets(context.Background(), MarketFilterParams{TagID: 12, EventID: 34}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListSeries(context.Background(), SeriesFilterParams{TagID: 56}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetComments(context.Background(), CommentFilterParams{EventID: 78}); err != nil {
		t.Fatal(err)
	}
}
