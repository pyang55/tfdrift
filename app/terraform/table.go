package terraform

import (
	"fmt"
	"os"
	"strconv"

	"tfdrift/log"

	v6table "github.com/jedib0t/go-pretty/v6/table"
)

func GenerateHTML(tsArray []*TerraformService) {
	f, err := os.Create("index.html")

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	f.WriteString(`
	<html>
	<head>
	  <title>Operations Weekly Summary</title>
	  <script src="https://www.kryogenix.org/code/browser/sorttable/sorttable.js"></script>
	  <script src="https://cdnjs.cloudflare.com/ajax/libs/jquery/3.3.1/jquery.min.js"></script>
	  <script src="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/js/bootstrap.min.js"></script>
	  <link rel="stylesheet" href="https://stackpath.bootstrapcdn.com/bootstrap/4.1.2/css/bootstrap.min.css" />
  
	  <style>
		body {
		  font-family: sans-serif;
		  background: linear-gradient(to left bottom, rgb(243, 244, 234) 0%, rgb(223, 220, 220) 100%);
		  margin: 0;
		  font-size: 16px;
		  padding: 16px;
		  color: rgb(80, 78, 78);
		}
  
		h1 {
		  margin-top: 0;
		  margin-bottom: 1.5rem;
		}
  
		h2 {
		  margin-top: 0;
		  margin-bottom: 0.875rem;
		}
  
		a {
		  color: #f1961d;
		}
		a:hover {
		  color: #d9881e;
		}
  
		hr {
		  margin: 1.5rem 0;
		}
  
		ul {
		  margin-top: 0;
		  margin-bottom: 1.5rem;
		}
  
		li {
		  margin-bottom: 4px;
		}
  
		table {
		  width: 100%;
		  display: table;
		  table-layout: fixed;
		  border-collapse: collapse;
		}
  
		table,
		th,
		td {
		  border: 1px solid black;
		}
  
		th,
		td {
		  padding: 10px;
		}
  
		th {
		  text-align: center;
		}
  
		td {
		  text-align: right;
		}
		td:first-of-type {
		  text-align: left;
		}
		table {
		  border-collapse: collapse;
		  color: #2e2e2e;
		  border: #a4a4a4;
		}
  
		table tr:hover {
		  background-color: #f2f2f2;
		}
  
		.details-row {
		  display: none;
		}
		
		.clickable-row {
		  cursor: pointer;
		}
		
		.clickable-row:hover {
		  background-color: #e8f4fd !important;
		}
	  </style>
	</head>
	<body>
	  <h1>Terraform Drift Report</h1>
	`)
	f.WriteString("<br />\n")
	f.WriteString("<button class=\"btn btn-secondary\" onclick=\"toggleAllRows()\">Expand/Collapse All</button>\n")
	f.WriteString("<hr />\n")
	// headers
	f.WriteString("<table class=\"sortable\">\n")
	f.WriteString("<thead>\n")
	f.WriteString("<tr>\n")
	f.WriteString("  <th>Project Name</th>\n")
	f.WriteString("  <th>Version</th>\n")
	f.WriteString("  <th>Add</th>\n")
	f.WriteString("  <th>Change</th>\n")
	f.WriteString("  <th>Delete</th>\n")
	f.WriteString("  <th>Information</th>\n")
	f.WriteString("</tr>\n")
	f.WriteString("</thead>\n")

	//body
	f.WriteString("<tbody>\n")
	t := 0
	for _, service := range tsArray {
		if service.Summary == "Drift detected for Plan." {
			// Create a safe ID by using the index if project name is empty/invalid
			safeId := service.ProjectName
			if safeId == "" || safeId == "." {
				safeId = fmt.Sprintf("project-%d", t)
			}
			f.WriteString(fmt.Sprintf("<tr id=\"%s\" class=\"clickable-row\" data-project=\"%s\">\n", safeId, safeId))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", service.ProjectName))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", service.TerraformVersion))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", strconv.Itoa(service.CountAdd)))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", strconv.Itoa(service.CountChange)))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", strconv.Itoa(service.CountDestroy)))
			f.WriteString(fmt.Sprintf("<td>%s</td>\n", service.Summary))
			f.WriteString("</tr>\n")

			// Hidden row containing the raw plan details
			f.WriteString(fmt.Sprintf("<tr id=\"%s-details\" class=\"details-row\">", safeId))
			f.WriteString("<td colspan=\"6\"><pre align=\"left\"><code>")
			fileName := fmt.Sprintf("/tmp/%s-tmp", service.ProjectName)
			b, err := os.ReadFile(fileName) // just pass the file name
			if err != nil {
				fmt.Print(err)
			}
			f.WriteString(string(b))
			f.WriteString("</code></pre></td>")
			f.WriteString("</tr>\n")
			t++
		}
	}
	f.WriteString("</tbody></table>\n")

	// JavaScript with proper event handling
	f.WriteString("<script type=\"text/javascript\">\n")
	f.WriteString("function toggleRow(id) {\n")
	f.WriteString("  console.log('toggleRow called with id:', id);\n")
	f.WriteString("  if (!id || id === '') {\n")
	f.WriteString("    console.error('Invalid id provided to toggleRow:', id);\n")
	f.WriteString("    return;\n")
	f.WriteString("  }\n")
	f.WriteString("  const detailsRow = document.querySelector('#' + id + '-details');\n")
	f.WriteString("  console.log('detailsRow found:', detailsRow);\n")
	f.WriteString("  if (detailsRow) {\n")
	f.WriteString("    const currentDisplay = detailsRow.style.display;\n")
	f.WriteString("    console.log('current display:', currentDisplay);\n")
	f.WriteString("    if (currentDisplay === 'none' || currentDisplay === '') {\n")
	f.WriteString("      detailsRow.style.display = 'table-row';\n")
	f.WriteString("      console.log('showing row');\n")
	f.WriteString("    } else {\n")
	f.WriteString("      detailsRow.style.display = 'none';\n")
	f.WriteString("      console.log('hiding row');\n")
	f.WriteString("    }\n")
	f.WriteString("  } else {\n")
	f.WriteString("    console.error('Details row not found for id:', id);\n")
	f.WriteString("  }\n")
	f.WriteString("}\n\n")

	f.WriteString("function toggleAllRows() {\n")
	f.WriteString("  const detailsRows = document.querySelectorAll('.details-row');\n")
	f.WriteString("  const anyVisible = Array.from(detailsRows).some(row => row.style.display === 'table-row');\n")
	f.WriteString("  detailsRows.forEach(row => {\n")
	f.WriteString("    row.style.display = anyVisible ? 'none' : 'table-row';\n")
	f.WriteString("  });\n")
	f.WriteString("}\n\n")

	f.WriteString("// Add event listeners after DOM is loaded\n")
	f.WriteString("document.addEventListener('DOMContentLoaded', function() {\n")
	f.WriteString("  console.log('DOM loaded, setting up click handlers');\n")
	f.WriteString("  const clickableRows = document.querySelectorAll('.clickable-row');\n")
	f.WriteString("  console.log('Found clickable rows:', clickableRows.length);\n")
	f.WriteString("  \n")
	f.WriteString("  clickableRows.forEach(function(row) {\n")
	f.WriteString("    const projectId = row.getAttribute('data-project') || row.id;\n")
	f.WriteString("    console.log('Setting up click handler for:', projectId);\n")
	f.WriteString("    row.addEventListener('click', function(e) {\n")
	f.WriteString("      console.log('Row clicked:', projectId);\n")
	f.WriteString("      toggleRow(projectId);\n")
	f.WriteString("    });\n")
	f.WriteString("  });\n")
	f.WriteString("});\n")
	f.WriteString("</script>\n")

	f.WriteString("</body>\n")
	f.WriteString("</html>\n")
}

// this is meant for stdout to allow for easier text manipluation
func PrettyTable(tsArray []*TerraformService) {

	t := v6table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(v6table.Row{"Project Name", "Version", "Add", "Change", "Delete", "Information"})
	drifts := 0
	for _, service := range tsArray {
		if service.Summary == "Drift detected for Plan." {
			t.AppendRows([]v6table.Row{{service.ProjectName, service.TerraformVersion, strconv.Itoa(service.CountAdd), strconv.Itoa(service.CountChange), strconv.Itoa(service.CountDestroy), service.Summary}})
			t.AppendSeparator()
			drifts++
		}
	}
	if drifts > 0 {
		t.SetStyle(v6table.StyleLight)
		t.Render()
	} else {
		fmt.Println("No Drift detected for the current infrastructure")
	}
	log.Debug("Sent Drift Report tables to stdout.")
}
