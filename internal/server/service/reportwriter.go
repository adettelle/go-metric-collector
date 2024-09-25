// сервисный слой
// отвечает за генерацию человекочитаемых отчетов о метриках
package service

import (
	"html/template"
	"io"
)

type Reporter interface {
	GetAllGaugeMetrics() (map[string]float64, error)
	GetAllCounterMetrics() (map[string]int64, error)
}

// функция использует любой объект, который имеет функции GetAllGaugeMetrics() и GetAllCounterMetrics()
// то есть который удовлетворяет Reporter'у
func WriteMetricsReport(rep Reporter, w io.Writer) error {

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

	GaugeMetric, err := rep.GetAllGaugeMetrics()
	if err != nil {
		return err
	}
	CounterMetric, err := rep.GetAllCounterMetrics()
	if err != nil {
		return err
	}
	m := tmlParams{
		Gauge:   GaugeMetric,
		Counter: CounterMetric,
	}
	err = t.Execute(w, m)
	if err != nil {
		return err
	}
	return nil
}
