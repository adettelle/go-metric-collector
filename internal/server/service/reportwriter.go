// сервисный слой
// отвечает за генерацию человекочитаемых отчетов о метриках
package service

import (
	"html/template"
	"io"
	"log"

	"github.com/adettelle/go-metric-collector/internal/storage/memstorage"
)

// type Reporter interface {
// 	GetAllGaugeMetrics() map[string]float64
// 	GetAllCounterMetrics() map[string]int64
// }

func WriteMetricsReport(ms *memstorage.MemStorage, w io.Writer) {

	const tmpl = `
<html>

	<body>
		<h1>Gauge metrics</h1>
    	<table> 
		{{range $key, $val := .Gauge}}
     		<tr>
				<td>{{$key}}</td>
				<td>{{$val}}</td> 
			</tr>
		{{end}}
		</table>

		<h1>Counter metrics</h1>
    	<table> 
		{{range $key, $val := .Counter}}
     		<tr>
				<td>{{$key}}</td>
				<td>{{$val}}</td> 
			</tr>
		{{end}}
		</table>
	</body>

</html>
	`
	t := template.Must(template.New("tmpl").Parse(tmpl))

	type tmlParams struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}

	m := tmlParams{
		Gauge:   ms.GetAllGaugeMetrics(),
		Counter: ms.GetAllCounterMetrics(),
	}
	err := t.Execute(w, m)
	if err != nil {
		log.Println("error:", err)
		return
	}
}
