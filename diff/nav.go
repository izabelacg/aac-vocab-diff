package diff

import (
	"database/sql"
	"fmt"
	"slices"
	"sort"
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

const pathArrow = " → "

const navEdgeQuery = `
SELECT source, btn_label, btn_message, target_name, target_value, vocab FROM (
SELECT
    r_page.name                  AS source,
    COALESCE(b.label, '')        AS btn_label,
    COALESCE(b.message, '')      AS btn_message,
    COALESCE(r_target.name, '')  AS target_name,
    COALESCE(ad0.value, '')      AS target_value,
    COALESCE(ad1.value, '')      AS vocab
FROM resources r_page
JOIN pages p                  ON p.resource_id     = r_page.id
JOIN button_box_instances bbi ON bbi.page_id       = p.id
JOIN button_boxes bb          ON bb.id             = bbi.button_box_id
JOIN button_box_cells bbc     ON bbc.button_box_id = bb.id
JOIN buttons b                ON b.resource_id     = bbc.resource_id
JOIN actions a                ON a.resource_id     = bbc.resource_id AND a.code IN (8, 9, 73)
LEFT JOIN action_data ad0     ON ad0.action_id = a.id AND ad0.key = 0
LEFT JOIN action_data ad1     ON ad1.action_id = a.id AND ad1.key = 1
LEFT JOIN resources r_target  ON r_target.rid = ad0.value

UNION ALL

-- buttons nested inside a button_set cell (type-5 resource in button_box_cells)
SELECT
    r_page.name                  AS source,
    COALESCE(b.label, '')        AS btn_label,
    COALESCE(b.message, '')      AS btn_message,
    COALESCE(r_target.name, '')  AS target_name,
    COALESCE(ad0.value, '')      AS target_value,
    COALESCE(ad1.value, '')      AS vocab
FROM resources r_page
JOIN pages p                  ON p.resource_id     = r_page.id
JOIN button_box_instances bbi ON bbi.page_id       = p.id
JOIN button_boxes bb          ON bb.id             = bbi.button_box_id
JOIN button_box_cells bbc     ON bbc.button_box_id = bb.id
JOIN button_sets bs           ON bs.resource_id    = bbc.resource_id
JOIN button_set_modifiers bsm ON bsm.button_set_id = bs.id
JOIN buttons b                ON b.id              = bsm.button_id
JOIN actions a                ON a.resource_id     = b.resource_id AND a.code IN (8, 9, 73)
LEFT JOIN action_data ad0     ON ad0.action_id = a.id AND ad0.key = 0
LEFT JOIN action_data ad1     ON ad1.action_id = a.id AND ad1.key = 1
LEFT JOIN resources r_target  ON r_target.rid = ad0.value
) ORDER BY source, CASE WHEN btn_label = '' THEN 1 ELSE 0 END, btn_label
`

func LoadPageNavGraph(dbPath string) (PageNavGraph, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", dbPath, err)
	}
	defer db.Close()

	return loadPageNavGraph(db)
}

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
		var source, btnLabel, btnMessage, targetName, targetValue, vocab string
		if err := rows.Scan(&source, &btnLabel, &btnMessage, &targetName, &targetValue, &vocab); err != nil {
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
		label := btnLabel
		if label == "" {
			if btnMessage == "" {
				label = "[unlabeled]"
			} else {
				label = fmt.Sprintf("<%s>", btnMessage)
			}
		}
		navGraph[source] = append(navGraph[source], NavEdge{Label: label, Dest: dest})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return navGraph, nil
}

func findRoots(g PageNavGraph, pages PageSet) []string {
	hasIncoming := map[string]bool{}
	for _, edges := range g {
		for _, e := range edges {
			hasIncoming[e.Dest] = true
		}
	}
	roots := []string{}
	for p := range pages {
		if !hasIncoming[p] {
			roots = append(roots, p)
		}
	}
	sort.Strings(roots)
	return roots
}

func AllShortestPaths(g PageNavGraph, pages PageSet) map[string]string {
	roots := findRoots(g, pages)
	if len(roots) == 0 {
		return map[string]string{}
	}
	reached := map[string]arrival{}
	toVisit := []string{}
	for _, r := range roots {
		reached[r] = arrival{from: "", button: ""}
		toVisit = append(toVisit, r)
	}

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
		if reached[page].from == "" {
			continue // root page — omit
		}

		labels := []string{}
		n := page
		for ; reached[n].from != ""; n = reached[n].from {
			labels = append(labels, reached[n].button)
		}
		slices.Reverse(labels)
		paths[page] = n + pathArrow + strings.Join(labels, pathArrow)
	}

	return paths
}
