package main
import (
 "fmt"
 "net/http"
 "reflect"
 "runtime"
 "sort"
 "app/internal/application"
 v1 "app/internal/routes/v1"
 "github.com/go-chi/chi/v5"
)
func name(v any) string { return runtime.FuncForPC(reflect.ValueOf(v).Pointer()).Name() }
func main() {
 for _, trust := range []bool{false, true} {
 r := v1.NewRouter(nil, nil, nil, &application.Application{}, trust)
 var lines []string
 err := chi.Walk(r, func(method, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
  n := "unknown"
  if f, ok := handler.(http.HandlerFunc); ok { n = name(f) }
  ms := ""
  for _, m := range middlewares { ms += " " + name(m) }
  lines = append(lines, fmt.Sprintf("%t %s %s %s%s", trust, method, route, n, ms))
  return nil
 })
 if err != nil { panic(err) }
 sort.Strings(lines)
 for _, line := range lines { fmt.Println(line) }
 }
}
