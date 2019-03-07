package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

//PreviewImage represents a preview image for a page
type PreviewImage struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Alt       string `json:"alt,omitempty"`
}

//PreviewVideo represents a preview video for a page **EXTRA CREDIT**
type PreviewVideo struct {
	URL       string `json:"url,omitempty"`
	SecureURL string `json:"secureURL,omitempty"`
	Type      string `json:"type,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

//PageSummary represents summary properties for a web page
type PageSummary struct {
	Type        string          `json:"type,omitempty"`
	URL         string          `json:"url,omitempty"`
	Title       string          `json:"title,omitempty"`
	SiteName    string          `json:"siteName,omitempty"`
	Description string          `json:"description,omitempty"`
	Author      string          `json:"author,omitempty"`
	Keywords    []string        `json:"keywords,omitempty"`
	Icon        *PreviewImage   `json:"icon,omitempty"`
	Images      []*PreviewImage `json:"images,omitempty"`
	Videos      []*PreviewVideo `json:"videos,omitempty"` //**EXTRA CREDIT**
}

//SummaryHandler handles requests for the page summary API.
//This API expects one query string parameter named `url`,
//which should contain a URL to a web page. It responds with
//a JSON-encoded PageSummary struct containing the page summary
//meta-data.
func SummaryHandler(w http.ResponseWriter, r *http.Request) {
	url := r.FormValue("url")
	body, err := fetchHTML(url)
	if err != nil {
		http.Error(w, "HTTP BAD REQUEST", 400)
		return
	}

	summary, err := extractSummary(url, body)
	defer body.Close()
	if err != nil {
		http.Error(w, "INTERNAL SERVER ERROR", 500)
		return
	}

	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	if err := enc.Encode(summary); err != nil {
		fmt.Fprintf(w, "error encoding struct into JSON: %v\n", err)
		return
	}

	// add appropriate headers when we know there are no errors

}

//fetchHTML fetches `pageURL` and returns the body stream or an error.
//Errors are returned if the response status code is an error (>=400),
//or if the content type indicates the URL is not an HTML page.
func fetchHTML(pageURL string) (io.ReadCloser, error) {
	res, err := http.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("Error, %s", err)
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("Error: %s", res.Status)
	}

	if !strings.HasPrefix(res.Header.Get("Content-Type"), "text/html") {
		return nil, fmt.Errorf("Only HTML content is accepted")
	}

	return res.Body, nil
}

//extractSummary tokenizes the `htmlStream` and populates a PageSummary
//struct with the page's summary meta-data.
func extractSummary(pageURL string, htmlStream io.ReadCloser) (*PageSummary, error) {
	summary := &PageSummary{}
	tokenizer := html.NewTokenizer(htmlStream)
	for {
		tokenType := tokenizer.Next()
		if tokenType == html.ErrorToken {
			if tokenizer.Err() == io.EOF {
				break
			} else {
				return nil, fmt.Errorf("error tokenizing HTML, %v", tokenizer.Err())
			}
		}

		token := tokenizer.Token()
		// only looking for start tags
		if tokenType == html.StartTagToken || tokenType == html.SelfClosingTagToken { // start and self closing tags
			if token.Data == "meta" { // meta tags
				prop, _ := getAttr(token, "property")
				name, _ := getAttr(token, "name")
				cont, _ := getAttr(token, "content")
				if prop == "og:type" {
					summary.Type = cont
				} else if prop == "og:url" {
					summary.URL = cont
				} else if prop == "og:title" {
					summary.Title = cont
				} else if prop == "og:site_name" {
					summary.SiteName = cont
				}

				if prop == "og:description" {
					summary.Description = cont
				} else if name == "description" && summary.Description == "" { // no open graph description
					summary.Description = cont
				}

				if name == "author" {
					summary.Author = cont
				}

				if name == "keywords" {
					if strings.IndexAny(cont, ",") != -1 { // remove all whitespace and split into slice
						summary.Keywords = strings.Split(strings.Replace(cont, " ", "", -1), ",")
					} else {
						summary.Keywords = []string{cont}
					}
				}

				if strings.HasPrefix(prop, "og:image") {
					if strings.HasPrefix(prop, "og:image:") { // is property of last PreviewImage made
						lastImage := summary.Images[len(summary.Images)-1]
						lastImage = buildImage(prop, pageURL, lastImage, cont)
					} else { // new image
						summary.Images = append(summary.Images, buildImage(prop, pageURL, &PreviewImage{}, cont))
					}
				}

				// for videos found in meta **EXTRA CREDIT**
				if strings.HasPrefix(prop, "og:video") {
					if strings.HasPrefix(prop, "og:video:") { // property of last PreviewVideo
						lastVideo := summary.Videos[len(summary.Videos)-1]
						lastVideo = buildVideo(prop, pageURL, lastVideo, cont)
					} else {
						summary.Videos = append(summary.Videos, buildVideo(prop, pageURL, &PreviewVideo{}, cont))
					}
				}
			}

			// if no Open Graph title
			if token.Data == "title" && summary.Title == "" {
				temp := tokenizer.Next()
				if temp == html.TextToken {
					summary.Title = tokenizer.Token().Data
				}
			}

			// link tags
			if token.Data == "link" {
				summary.Icon = buildIcon(token, pageURL)
			}
		}
		// end after we have parsed head tag
		if tokenType == html.EndTagToken && token.Data == "head" {
			break
		}
	}
	return summary, nil
}

// getAttr takes in an html token from a tokenizer along with a desired attribute string
// and searches/returns the val of that attribute if found. Otherwise it returns an empty string
// as well as a custom error
func getAttr(token html.Token, attr string) (string, error) {
	for _, a := range token.Attr {
		if a.Key == attr { // found it
			return a.Val, nil
		}
	}
	return "", errors.New("Invalid or nonexistent Attribute") // not found
}

// resolveURL takes in a base URL string and relative URL string and returns an absolute URL string
// that can locate the relative resource
func resolveURL(base string, loc string) string {
	// convert both strings into URLs
	bURL, _ := url.Parse(base)
	lURL, _ := url.Parse(loc)

	// make an absolute path to resource, will ignore if already absolute
	return bURL.ResolveReference(lURL).String()
}

// buildIcon takes in a token and pageURL to construct and return a
// PreviewImage struct using the attributes in the token and resolves
// any relative URLs using the pageURL
func buildIcon(token html.Token, pageURL string) *PreviewImage {
	rel, _ := getAttr(token, "rel")
	temp := &PreviewImage{}
	if rel == "icon" {
		href, _ := getAttr(token, "href")
		typ, _ := getAttr(token, "type")
		alt, _ := getAttr(token, "alt")
		sizes, _ := getAttr(token, "sizes")

		if !strings.HasPrefix(href, "http") {
			href = resolveURL(pageURL, href)
		}
		temp.URL = href
		temp.Alt = alt
		temp.Type = typ

		if sizes != "any" && sizes != "" {
			sizeSlice := strings.Split(sizes, "x")
			temp.Height, _ = strconv.Atoi(sizeSlice[0])
			temp.Width, _ = strconv.Atoi(sizeSlice[1])
		}
	}
	return temp
}

// buildImage takes in a property string, pageURL, PreviewImage struct and content string to add to the
// PreviewImage using the content in the cont string based on what property it belongs to and resolves
// any relative URLs using the pageURL
func buildImage(prop string, pageURL string, image *PreviewImage, cont string) *PreviewImage {
	if prop == "og:image" {
		if !strings.HasPrefix(cont, "http") { // not absolute path
			cont = resolveURL(pageURL, cont)
		}
		image.URL = cont
	}

	if prop == "og:image:secure_url" {
		image.SecureURL = resolveURL(pageURL, cont)
	} else if prop == "og:image:type" {
		image.Type = cont
	} else if prop == "og:image:width" {
		image.Width, _ = strconv.Atoi(cont)
	} else if prop == "og:image:height" {
		image.Height, _ = strconv.Atoi(cont)
	} else if prop == "og:image:alt" {
		image.Alt = cont
	}
	return image
}

// function that takes in a property string, pageURL, pointer to PreviewVideo and a content string
// to build a PreviewVideo object and return it
func buildVideo(prop string, pageURL string, video *PreviewVideo, cont string) *PreviewVideo {
	if prop == "og:video" {
		if !strings.HasPrefix(cont, "http") {
			cont = resolveURL(pageURL, cont)
		}
		video.URL = cont
	}

	if prop == "og:video:secure_url" {
		video.SecureURL = resolveURL(pageURL, cont)
	} else if prop == "og:video:type" {
		video.Type = cont
	} else if prop == "og:image:width" {
		video.Width, _ = strconv.Atoi(cont)
	} else if prop == "og:image:height" {
		video.Height, _ = strconv.Atoi(cont)
	}
	return video
}
