package webui

import (
	"fmt"
	"html/template"
	"net/http"
)

var indexTemplate = template.Must(template.New("index").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>AI Content Engine Admin</title>
  <style>
    body { font-family: system-ui, sans-serif; margin: 2rem auto; max-width: 760px; padding: 0 1rem; color: #111827; background: #f9fafb; }
    h1 { margin-bottom: .25rem; }
    p { color: #4b5563; }
    .card { background: #fff; border: 1px solid #e5e7eb; border-radius: 10px; padding: 1rem; margin-top: 1rem; }
    a { color: #1d4ed8; text-decoration: none; }
    a:hover { text-decoration: underline; }
    ul { margin: .5rem 0 0 1.25rem; }
  </style>
</head>
<body>
  <h1>AI Content Engine</h1>
  <p>Basic admin web UI</p>

  <div class="card">
    <strong>Quick links</strong>
    <ul>
      <li><a href="/health">Health check</a></li>
      <li><a href="/admin/drafts?status=pending&limit=20">Pending drafts</a></li>
      <li><a href="/admin/drafts?status=approved&limit=20">Approved drafts</a></li>
      <li><a href="/admin/drafts?status=rejected&limit=20">Rejected drafts</a></li>
      <li><a href="/admin/drafts?status=posted&limit=20">Posted drafts</a></li>
    </ul>
  </div>
</body>
</html>`))

// Register wires basic web-ui routes.
func Register(mux *http.ServeMux) error {
	if mux == nil {
		return fmt.Errorf("mux is nil")
	}
	mux.HandleFunc("/", handleIndex)
	return nil
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = indexTemplate.Execute(w, nil)
}
