package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/net/proxy"
	"h12.io/socks"
)

var (
	tip          string
	path         string
	threadsInput *widget.Entry
	pencere      fyne.Window
)

func main() {
	uygulama := app.New()
	pencere = uygulama.NewWindow("Proxy Checker")
	pencere.Resize(fyne.NewSize(400, 200))
	pencere.CenterOnScreen()
	threadsLabel := widget.NewLabel("Threads : ")
	threadsInput = widget.NewEntry()
	threadsInput.SetText("100")
	container := container.NewVBox(
		widget.NewSelect([]string{"HTTP", "SOCKS4", "SOCKS5"}, func(s string) {
			tip = s
		}),
		threadsLabel,
		threadsInput,

		widget.NewButton("Proxyleri Seç", func() {
			dialog := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
				if err == nil && uc == nil {
					return
				}
				if err != nil {
					dialog.ShowError(err, pencere)
					return
				}
				path = uc.URI().String()
			}, pencere)
			dialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))
			dialog.Show()

		}),
		widget.NewButton("Başlat", func() {
			proxyReader(strings.Split(path, "file://")[1])
			dialog := dialog.NewInformation("Tarama Bitti", "Tarama bitti programı kapatabilirsiniz", pencere)
			dialog.Show()
		}),
	)

	pencere.SetContent(container)
	pencere.ShowAndRun()
}

// credits : https://gist.github.com/pkulak/93336af9bb9c7207d592
func proxyReader(path string) {
	workQueue := make(chan string)
	complete := make(chan bool)
	go func() {
		file, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			workQueue <- scanner.Text()
		}

		close(workQueue)
	}()
	threadsInt, _ := strconv.Atoi(threadsInput.Text)
	for i := 0; i < threadsInt; i++ {
		go checker(workQueue, complete)
	}

	for i := 0; i < threadsInt; i++ {
		<-complete

	}
}

func checker(queue chan string, complete chan bool) {
	for line := range queue {
		split := strings.Split(line, ":")
		if tip == "HTTP" {
			proxyurl, _ := url.Parse(fmt.Sprintf("http://%s:%s", split[0], split[1]))
			httpclient := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyurl),
				},
			}
			req, err := httpclient.Get("https://google.com/")
			if err != nil {
				continue
			} else {
				dosya, _ := os.OpenFile("http_live_proxies.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
				dosya.WriteString(line + "\n")
				dosya.Close()
			}
			req.Body.Close()
		} else if tip == "SOCKS5" {
			_, err := proxy.SOCKS5("tcp", line, nil, proxy.Direct)
			if err != nil {
				continue
			} else {
				dosya, _ := os.OpenFile("socks5_live_proxies.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
				dosya.WriteString(line + "\n")
				dosya.Close()
			}

		} else {
			dialer := socks.Dial(fmt.Sprintf("socks4://%s?timeout=4s", line))
			httpclient := &http.Client{
				Transport: &http.Transport{
					Dial: dialer,
				},
			}
			req, err := httpclient.Get("https://google.com/")
			if err != nil {
				continue
			} else {
				dosya, _ := os.OpenFile("socks4_live_proxies.txt", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0755)
				dosya.WriteString(line + "\n")
				dosya.Close()
			}
			req.Body.Close()
		}
	}
	complete <- true
}
