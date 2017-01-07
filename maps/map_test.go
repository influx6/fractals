package maps_test

import (
	"fmt"
	"testing"

	"github.com/influx6/fractals"
	"github.com/influx6/fractals/maps"
)

// succeedMark is the Unicode codepoint for a check mark.
const succeedMark = "\u2713"

// failedMark is the Unicode codepoint for an X mark.
const failedMark = "\u2717"

func TestMapFind(t *testing.T) {
	tree := map[string]interface{}{
		"name":   "wonder",
		"prices": []int{1, 500, 433, 5000, 320},
		"documents": []map[string]interface{}{
			{
				"metrics": map[string]string{
					"name": "bunny",
				},
				"data": []float64{1.435, 5.00, 43.3, 50.00, 3.20},
			},
		},
		"meta": map[string]interface{}{
			"mark":     300,
			"desc":     "weather bill of the year",
			"sentries": []string{"bh-300", "vh-10", "bl-30"},
		},
	}

	nameFinder := maps.Find("name")
	find(t, nameFinder, "name", tree)

	pricesFinder := maps.Find("prices.2")
	find(t, pricesFinder, "prices.2", tree)

	metaFinder := maps.Find("meta.desc")
	find(t, metaFinder, "meta.desc", tree)

	docFinder := maps.Find("documents.0.metrics.name")
	find(t, docFinder, "documents.0.metrics.name", tree)

	dataFinder := maps.Find("documents.0.data.1")
	find(t, dataFinder, "documents.0.data.1", tree)
}

func TestMapSave(t *testing.T) {
	tree := map[string]interface{}{
		"name":   "wonder",
		"prices": []int{1, 500, 433, 5000, 320},
		"documents": []map[string]interface{}{
			{
				"metrics": map[string]string{
					"name": "bunny",
				},
				"data": []float64{1.435, 5.00, 43.3, 50.00, 3.20},
			},
		},
		"meta": map[string]interface{}{
			"mark":     300,
			"desc":     "weather bill of the year",
			"sentries": []string{"bh-300", "vh-10", "bl-30"},
		},
	}

	docFinder := maps.Find("documents.0.metrics.name")
	find(t, docFinder, "documents.0.metrics.name", tree)

	docSetter := maps.Save("documents.0.metrics.name", "tord")
	set(t, docSetter, "documents.0.metrics.name", tree)

	find(t, docFinder, "documents.0.metrics.name", tree)

	nameFinder := maps.Find("name")
	find(t, nameFinder, "name", tree)

	nameSetter := maps.Save("name", "star-trek")
	set(t, nameSetter, "name", tree)
	find(t, nameFinder, "name", tree)
}

func find(t *testing.T, handler fractals.Handler, key string, target interface{}) {
	value, err := handler(nil, nil, target)
	if err != nil {
		fatalFailed(t, "Should have found key path from the provided target: ", err)
	}

	logPassed(t, "Should have found key path from the provided target:  key[%s] and Value[%#v]", key, value)
}

func set(t *testing.T, handler fractals.Handler, key string, target interface{}) {
	value, err := handler(nil, nil, target)
	if err != nil {
		fatalFailed(t, "Should have set key path with value[%#v] from the provided target: ", err, target)
	}

	logPassed(t, "Should have set key path with key[%s]  and Value[%#v]", key, value)
}

func logPassed(t *testing.T, msg string, data ...interface{}) {
	t.Logf("%s %s", fmt.Sprintf(msg, data...), succeedMark)
}

func fatalFailed(t *testing.T, msg string, data ...interface{}) {
	t.Fatalf("%s %s", fmt.Sprintf(msg, data...), failedMark)
}
