package controllers

import "net/http"

import "github.com/gin-gonic/gin"

// Source describes one RSS feed used for aggregation, along with a
// commonly cited characterization of its editorial/political leaning.
//
// These labels are indicative simplifications of a much more nuanced
// reality (ownership, editorial line, and journalists' own views can all
// diverge), drawn from widely shared media-literacy characterizations of
// the French press landscape. They are provided for transparency so
// readers can factor source leaning into how they read an article, not as
// an authoritative judgment.
type Source struct {
	Name             string `json:"name"`
	URL              string `json:"-"`
	PoliticalLeaning string `json:"political_leaning"`
}

var feedSources = []Source{
	{
		Name:             "France Info",
		URL:              "https://www.franceinfo.fr/france.rss",
		PoliticalLeaning: "Service public - ligne éditoriale neutre",
	},
	{
		Name:             "Le Figaro",
		URL:              "https://www.lefigaro.fr/rss/figaro_actualites.xml",
		PoliticalLeaning: "Droite",
	},
	{
		Name:             "Le Monde",
		URL:              "https://www.lemonde.fr/rss/une.xml",
		PoliticalLeaning: "Centre gauche",
	},
	{
		Name:             "Libération",
		URL:              "https://www.liberation.fr/arc/outboundfeeds/rss/?outputType=xml",
		PoliticalLeaning: "Gauche",
	},
	{
		Name:             "20 Minutes",
		URL:              "https://www.20minutes.fr/feeds/rss-une.xml",
		PoliticalLeaning: "Généraliste - ligne éditoriale peu marquée",
	},
	{
		Name:             "L'Express",
		URL:              "https://www.lexpress.fr/arc/outboundfeeds/rss/alaune.xml",
		PoliticalLeaning: "Centre droit",
	},
	{
		Name:             "BFM TV",
		URL:              "https://www.bfmtv.com/rss/news-24-7/",
		PoliticalLeaning: "Généraliste - ligne éditoriale souvent qualifiée de centre droit",
	},
	{
		Name:             "France 24",
		URL:              "https://www.france24.com/fr/rss",
		PoliticalLeaning: "Service public - ligne éditoriale neutre",
	},
	{
		Name:             "RFI",
		URL:              "https://www.rfi.fr/fr/rss",
		PoliticalLeaning: "Service public - ligne éditoriale neutre",
	},
	{
		Name:             "Ouest-France",
		URL:              "https://www.ouest-france.fr/rss-en-continu.xml",
		PoliticalLeaning: "Centre - presse régionale, ligne modérée",
	},
}

// politicalLeaningBySource indexes feedSources by name for quick lookup
// when enriching articles read back from the database.
var politicalLeaningBySource = func() map[string]string {
	index := make(map[string]string, len(feedSources))
	for _, source := range feedSources {
		index[source.Name] = source.PoliticalLeaning
	}
	return index
}()

// GetSources returns the list of aggregated sources along with their
// indicative political leaning, so a client can display it as a legend.
func GetSources(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"sources": feedSources,
		"note":    "Ces orientations sont des caractérisations indicatives et simplifiées, à but de transparence, et ne constituent pas un jugement définitif.",
	})
}
