package diff

import (
	"database/sql"
	"slices"
	"strings"
)

type arrival struct {
	from   string // the page you were on
	button string // button label pressed on 'from' to arrive here
}

type NavEdge struct {
	Label string
	Dest  string
}

type PageNavGraph map[string][]NavEdge

const (
	navPathRootPage = "Home"
	pathArrow       = " → "
)

const navEdgeQuery = `
SELECT
    r_page.name                  AS source,
    COALESCE(b.label, '')        AS btn_label,
    COALESCE(r_target.name, '')  AS target_name,
    COALESCE(ad0.value, '')      AS target_value,
    COALESCE(ad1.value, '')      AS vocab
FROM resources r_page
JOIN pages p                  ON p.resource_id     = r_page.id
JOIN button_box_instances bbi ON bbi.page_id       = p.id
JOIN button_boxes bb          ON bb.id             = bbi.button_box_id
JOIN button_box_cells bbc     ON bbc.button_box_id = bb.id
JOIN buttons b                ON b.resource_id     = bbc.resource_id
JOIN actions a                ON a.resource_id     = bbc.resource_id AND a.code = 9
LEFT JOIN action_data ad0     ON ad0.action_id = a.id AND ad0.key = 0
LEFT JOIN action_data ad1     ON ad1.action_id = a.id AND ad1.key = 1
LEFT JOIN resources r_target  ON r_target.rid = ad0.value
ORDER BY r_page.name, b.label
`

func loadPageNavGraph(db *sql.DB) (PageNavGraph, error) {
	rows, err := db.Query(navEdgeQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// seenDest prevents duplicate (source, dest) edges; first label wins
	// because rows arrive ORDER BY source, btn_label.
	seenDest := map[string]map[string]struct{}{} // source → set of dest

	navGraph := make(PageNavGraph)
	for rows.Next() {
		var source, btnLabel, targetName, targetValue, vocab string
		if err := rows.Scan(&source, &btnLabel, &targetName, &targetValue, &vocab); err != nil {
			return nil, err
		}
		dest := navigateDestination(targetName, targetValue, vocab)

		if seenDest[source] == nil {
			seenDest[source] = make(map[string]struct{})
		}
		if _, seen := seenDest[source][dest]; seen {
			continue // already have an edge from source to dest
		}
		seenDest[source][dest] = struct{}{}
		navGraph[source] = append(navGraph[source], NavEdge{Label: btnLabel, Dest: dest})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return navGraph, nil
}

func AllShortestPathsFromHome(g PageNavGraph, pages PageSet) map[string]string {
	if _, ok := pages[navPathRootPage]; !ok {
		return nil
	}

	reached := map[string]arrival{navPathRootPage: {from: "", button: ""}} // page → how we got there
	toVisit := []string{navPathRootPage}

	for len(toVisit) > 0 {
		cur := toVisit[0]
		toVisit = toVisit[1:]

		for _, edge := range g[cur] {
			if _, visited := reached[edge.Dest]; !visited {
				reached[edge.Dest] = arrival{from: cur, button: edge.Label}
				toVisit = append(toVisit, edge.Dest)
			}
		}
	}

	paths := map[string]string{}
	for page := range reached {
		if page == navPathRootPage {
			paths[page] = navPathRootPage
			continue
		}
		labels := []string{}
		for n := page; reached[n].from != ""; n = reached[n].from {
			labels = append(labels, reached[n].button)
		}
		slices.Reverse(labels)
		paths[page] = navPathRootPage + pathArrow + strings.Join(labels, pathArrow)
	}

	return paths
}
