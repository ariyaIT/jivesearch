package instant

import (
	"net/http"

	"github.com/jivesearch/jivesearch/instant/contributors"
	"github.com/jivesearch/jivesearch/log"
	"github.com/jivesearch/jivesearch/wikipedia"
	"golang.org/x/text/language"
)

// WikiData is an instant answer
type WikiData struct {
	wikipedia.Fetcher
	Answer
}

func (w *WikiData) setQuery(r *http.Request, qv string) answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *WikiData) setUserAgent(r *http.Request) answerer {
	return w
}

func (w *WikiData) setType() answerer {
	w.Type = "wikidata"
	return w
}

func (w *WikiData) setContributors() answerer {
	w.Contributors = contributors.Load(
		[]string{
			"brentadamson",
		},
	)
	return w
}

// trigger words
// age ---> for "how old is x?" we need to change our triggerfuncs to just a regex
const age = "age"
const howOldIs = "how old is"

// birthday
const birthday = "birthday"
const born = "born"

// death
const death = "death"
const died = "died"

// height
const height = "height"
const howTallis = "how tall is"
const howTallwas = "how tall was"

func (w *WikiData) setTriggers() answerer {
	w.triggers = []string{
		age, howOldIs,
		birthday, born,
		death, died,
		howTallis, howTallwas, height,
	}
	return w
}

func (w *WikiData) setTriggerFuncs() answerer {
	w.triggerFuncs = []triggerFunc{
		startsWith, endsWith,
	}
	return w
}

// Birthday is a person's date of birth
type Birthday struct {
	Birthday wikipedia.DateTime `json:"birthday,omitempty"`
	Alive    bool               `json:"alive,omitempty"`
}

// Death is a person's date of death
// TODO: add place of death, cause, etc.
type Death struct {
	Death wikipedia.DateTime `json:"death,omitempty"`
}

// Age is a person's current age (in years) or age when they died
type Age struct {
	Birthday `json:"birthday,omitempty"`
	Death    `json:"death,omitempty"`
}

// TODO: Return the Title (and perhaps Image???) as
// confirmation that we fetched the right asset.
func (w *WikiData) setSolution() answerer {
	item, err := w.Fetch(w.remainder, language.English)
	if err != nil {
		log.Info.Println(err) // errors should be returned to caller, not handled here
	}

	switch w.triggerWord {
	case age, howOldIs, birthday, born:
		if len(item.Birthday) == 0 {
			return w
		}

		b := Birthday{item.Birthday[0], true}
		if len(item.Death) > 0 {
			b.Alive = false
		}

		if w.triggerWord == "age" || w.triggerWord == "how old is" {
			a := Age{
				Birthday: b,
			}

			if len(item.Death) > 0 {
				a.Death = Death{item.Death[0]}
			}

			w.Solution.Raw = a

			return w
		}

		w.Solution.Raw = b
	case death, died:
		if len(item.Death) > 0 {
			w.Solution.Raw = Death{item.Death[0]}
		}
	case howTallis, howTallwas, height:
		if len(item.Height) == 0 {
			return w
		}

		w.Solution.Raw = item.Height
	}

	return w
}

func (w *WikiData) setCache() answerer {
	w.Cache = true
	return w
}

func (w *WikiData) tests() []test {
	typ := "wikidata"

	contrib := contributors.Load([]string{"brentadamson"})

	tests := []test{
		{
			query: "Bob Marley age",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Raw: Age{
						Birthday: Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1945-02-06T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
							Alive: false,
						},
						Death: Death{
							Death: wikipedia.DateTime{
								Value:    "1981-05-11T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "Jimi hendrix birthday",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Raw: Birthday{
						Birthday: wikipedia.DateTime{
							Value:    "1942-11-27T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
						Alive: false,
					},
					Cache: true,
				},
			},
		},
		{
			query: "death jimi hendrix",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Raw: Death{
						Death: wikipedia.DateTime{
							Value:    "1970-09-18T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Cache: true,
				},
			},
		},
		{
			query: "shaquille o'neal height",
			expected: []Solution{
				{
					Type:         typ,
					Triggered:    true,
					Contributors: contrib,
					Raw: []wikipedia.Quantity{
						{
							Amount: "2.16",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}

// mock Wikidata Fetcher
type mockFetcher struct{}

func (mf *mockFetcher) Fetch(query string, lang language.Tag) (*wikipedia.Item, error) {
	switch query {
	case "bob marley":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Birthday: []wikipedia.DateTime{
						{
							Value:    "1945-02-06T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Death: []wikipedia.DateTime{
						{
							Value:    "1981-05-11T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		}, nil
	case "jimi hendrix":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Birthday: []wikipedia.DateTime{
						{
							Value:    "1942-11-27T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
					Death: []wikipedia.DateTime{
						{
							Value:    "1970-09-18T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		}, nil

	case "shaquille o'neal":
		return &wikipedia.Item{
			Wikidata: &wikipedia.Wikidata{
				Claims: &wikipedia.Claims{
					Height: []wikipedia.Quantity{
						{
							Amount: "2.16",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
				},
			},
		}, nil

	}

	return &wikipedia.Item{}, nil

}

func (mf *mockFetcher) Setup() error {
	return nil
}
