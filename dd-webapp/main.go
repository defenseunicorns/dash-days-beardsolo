package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

type OscalComponent struct {
	ComponentDefinition struct {
		UUID     string `yaml:"uuid"`
		Metadata struct {
			Version      string `yaml:"version"`
			LastModified string `yaml:"last-modified"`
			OscalVersion string `yaml:"oscal-version"`
			Title        string `yaml:"title"`
			Parties      []struct {
				Type  string `yaml:"type"`
				Name  string `yaml:"name"`
				UUID  string `yaml:"uuid"`
				Links []struct {
					Rel  string `yaml:"rel"`
					Href string `yaml:"href"`
				} `yaml:"links"`
			} `yaml:"parties"`
		} `yaml:"metadata"`
		Components []struct {
			UUID             string `yaml:"uuid"`
			Title            string `yaml:"title"`
			Description      string `yaml:"description"`
			Type             string `yaml:"type"`
			Purpose          string `yaml:"purpose"`
			ResponsibleRoles []struct {
				RoleID     string   `yaml:"role-id"`
				PartyUUIDs []string `yaml:"party-uuids"`
			} `yaml:"responsible-roles"`
			ControlImplementations []struct {
				Source                  string `yaml:"source"`
				Description             string `yaml:"description"`
				UUID                    string `yaml:"uuid"`
				ImplementedRequirements []struct {
					UUID        string `yaml:"uuid"`
					ControlID   string `yaml:"control-id"`
					Description string `yaml:"description"`
				} `yaml:"implemented-requirements"`
			} `yaml:"control-implementations"`
		} `yaml:"components"`
	} `yaml:"component-definition"`
}

func main() {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/public", "./public")

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("File upload error: %s", err.Error()))
			return
		}

		filename := filepath.Base(file.Filename)
		tempFilePath := filepath.Join("uploads", filename)

		// Save the file to the server's local storage
		err = c.SaveUploadedFile(file, tempFilePath)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save file: %s", err.Error()))
			return
		}

		// Process the CSV file
		oscalComponent, err := processCSV(tempFilePath)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to process file: %s", err.Error()))
			return
		}

		// Convert the structure to YAML
		yamlData, err := yaml.Marshal(&oscalComponent)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal YAML: %s", err.Error()))
			return
		}

		// Write the YAML data to a file or send it as a response
		fmt.Println(string(yamlData))
		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded and processed successfully.", file.Filename))

		// Call the function to save YAML data to file
		err = saveYAMLToFile(yamlData)
		if err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprintf("Failed to save YAML to file: %s", err.Error()))
			return
		}

		// Respond with YAML content
		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("File %s uploaded and processed successfully.", file.Filename),
			"yamlData": string(yamlData),
		})
	})

	// Define a new route for downloading the YAML file
	r.GET("/download", func(c *gin.Context) {
		// Specify the file path
		filePath := "oscal-component.yaml"

		// Check if the file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.String(http.StatusNotFound, fmt.Sprintf("File %s does not exist.", filePath))
			return
		}

		// Serve the file for download
		c.File(filePath)
	})

	r.Run(":8080")
}

func processCSV(filePath string) (OscalComponent, error) {
	var oscalComponent OscalComponent
	// Initialize oscalComponent with predefined values

	oscalComponent.ComponentDefinition.UUID = uuid.New().String()
	oscalComponent.ComponentDefinition.Metadata.Version = "0.0.1"
	oscalComponent.ComponentDefinition.Metadata.LastModified = time.Now().Format(time.RFC3339)
	oscalComponent.ComponentDefinition.Metadata.OscalVersion = "1.0.4"
	oscalComponent.ComponentDefinition.Metadata.Title = "DUBBD"
	oscalComponent.ComponentDefinition.Metadata.Parties = []struct {
		Type  string `yaml:"type"`
		Name  string `yaml:"name"`
		UUID  string `yaml:"uuid"`
		Links []struct {
			Rel  string `yaml:"rel"`
			Href string `yaml:"href"`
		} `yaml:"links"`
	}{
		{
			Type: "organization",
			Name: "Defense Unicorns",
			UUID: uuid.New().String(),
			Links: []struct {
				Rel  string `yaml:"rel"`
				Href string `yaml:"href"`
			}{
				{Rel: "website", Href: "https://defenseunicorns.com"},
			},
		},
	}

	file, err := os.Open(filePath)
	if err != nil {
		return oscalComponent, err
	}
	defer file.Close()

	csvReader := csv.NewReader(file)
	_, err = csvReader.Read() // skip the header
	if err != nil {
		return oscalComponent, err
	}

	records, err := csvReader.ReadAll()
	if err != nil {
		return oscalComponent, err
	}

	for _, record := range records {
		componentUUID := uuid.New().String()
		requirementUUID := uuid.New().String()

		oscalComponent.ComponentDefinition.Components = append(oscalComponent.ComponentDefinition.Components, struct {
			UUID             string `yaml:"uuid"`
			Title            string `yaml:"title"`
			Description      string `yaml:"description"`
			Type             string `yaml:"type"`
			Purpose          string `yaml:"purpose"`
			ResponsibleRoles []struct {
				RoleID     string   `yaml:"role-id"`
				PartyUUIDs []string `yaml:"party-uuids"`
			} `yaml:"responsible-roles"`
			ControlImplementations []struct {
				Source                  string `yaml:"source"`
				Description             string `yaml:"description"`
				UUID                    string `yaml:"uuid"`
				ImplementedRequirements []struct {
					UUID        string `yaml:"uuid"`
					ControlID   string `yaml:"control-id"`
					Description string `yaml:"description"`
				} `yaml:"implemented-requirements"`
			} `yaml:"control-implementations"`
		}{
			UUID:        componentUUID,
			Title:       record[1], // Component Name
			Description: record[2], // Control Description
			Type:        "software",
			Purpose:     "Purpose of the component", // Replace with actual purpose
			ResponsibleRoles: []struct {
				RoleID     string   `yaml:"role-id"`
				PartyUUIDs []string `yaml:"party-uuids"`
			}{
				{RoleID: "provider", PartyUUIDs: []string{uuid.New().String()}},
			},
			ControlImplementations: []struct {
				Source                  string `yaml:"source"`
				Description             string `yaml:"description"`
				UUID                    string `yaml:"uuid"`
				ImplementedRequirements []struct {
					UUID        string `yaml:"uuid"`
					ControlID   string `yaml:"control-id"`
					Description string `yaml:"description"`
				} `yaml:"implemented-requirements"`
			}{
				{
					Source:      "https://raw.githubusercontent.com/usnistgov/oscal-content/master/nist.gov/SP800-53/rev5/json/NIST_SP-800-53_rev5_catalog.json",
					Description: "Controls implemented by " + record[1] + " for inheritance by applications",
					UUID:        uuid.New().String(),
					ImplementedRequirements: []struct {
						UUID        string `yaml:"uuid"`
						ControlID   string `yaml:"control-id"`
						Description string `yaml:"description"`
					}{
						{
							UUID:        requirementUUID,
							ControlID:   record[0], // Control Acronym
							Description: record[2], // Control Description
						},
					},
				},
			},
		})
	}

	return oscalComponent, nil

}

func saveYAMLToFile(yamlData []byte) error {
	filePath := "oscal-component.yaml"
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("Failed to create file %s: %v", filePath, err)
	}
	defer file.Close()

	_, err = file.Write(yamlData)
	if err != nil {
		return fmt.Errorf("Failed to write data to file %s: %v", filePath, err)
	}

	return nil
}
