// Package sources is the single source of truth for the RSS feeds the
// application aggregates. It is imported both by the controllers (to fetch
// feeds and expose them over the API) and by the db package (to seed the
// "sources" table), so the two never drift apart.
package sources

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

var FeedSources = []Source{
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
