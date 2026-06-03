package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

const PORT = "8080"
const AUTHOR = "Kacper Kantarowicz"

// Struktury do odczytania JSONa z wttr.in
type WttrResponse struct {
	CurrentCondition []struct {
		TempC         string `json:"temp_C"`
		WindspeedKmph string `json:"windspeedKmph"`
		WeatherDesc   []struct {
			Value string `json:"value"`
		} `json:"weatherDesc"`
	} `json:"current_condition"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-health" {
		resp, err := http.Get("http://localhost:" + PORT + "/")
		if err != nil || resp.StatusCode != 200 {
			os.Exit(1) // Healthcheck niezaliczony
		}
		os.Exit(0) // Healthcheck zaliczony
	}

	// logi po uruchomieniu
	fmt.Printf("data uruchomienia: %s\n", time.Now().Format(time.RFC3339))
	fmt.Printf("autor: %s\n", AUTHOR)
	fmt.Printf("serwer nasłuchuje na porcie TCP: %s\n", PORT)

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/weather", weatherHandler)

	http.ListenAndServe(":"+PORT, nil)
}

// Interfejs
func indexHandler(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="pl">
<head>
    <meta charset="UTF-8"><title>Apka Pogodowa</title>
</head>
<body>
    <h2>Wybierz lokalizację</h2>
    <select id="loc">
		<option value="Lublin">Lublin</option>
        <option value="Warsaw">Warszawa</option>
        <option value="London">Londyn</option>
        <option value="Tokyo">Tokio</option>
    </select>
    <button onclick="getW()">Pokaż pogodę</button>
    <div id="res"></div>
    <script>
        async function getW() {
            document.getElementById('res').innerText = "Ładowanie...";
            try {
                let r = await fetch('/api/weather?city=' + document.getElementById('loc').value);
                let d = await r.json();
                document.getElementById('res').innerText = 'Temp: '+d.temp+'°C | Wiatr: '+d.wind+' km/h | '+d.cond;
            } catch(e) { document.getElementById('res').innerText = "Błąd."; }
        }
    </script>
</body></html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, html)
}

// Pobieranie pogody
func weatherHandler(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Warsaw"
	}
	resp, err := http.Get("https://wttr.in/" + city + "?format=j1")
	if err != nil {
		http.Error(w, "Error", 500)
		return
	}
	defer resp.Body.Close()

	var wttr WttrResponse
	json.NewDecoder(resp.Body).Decode(&wttr)

	if len(wttr.CurrentCondition) > 0 {
		c := wttr.CurrentCondition[0]
		desc := ""
		if len(c.WeatherDesc) > 0 {
			desc = c.WeatherDesc[0].Value
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"temp":"%s", "wind":"%s", "cond":"%s"}`, c.TempC, c.WindspeedKmph, desc)
	} else {
		http.Error(w, "Not found", 404)
	}
}