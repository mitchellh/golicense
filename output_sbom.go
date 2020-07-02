package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/cloudentity/golicense/config"
	"github.com/cloudentity/golicense/license"
	"github.com/cloudentity/golicense/module"
	"github.com/package-url/packageurl-go"
	"github.com/pkg/errors"
)

type BOM struct {
	XMLName      xml.Name    `xml:"bom"`
	XMLNs        string      `xml:"xmlns,attr"`
	Version      int         `xml:"version,attr"`
	SerialNumber string      `xml:"serialNumber,attr"`
	Components   []Component `xml:"components>component"`
}

type Component struct {
	XMLName  xml.Name  `xml:"component"`
	Type     string    `xml:"type,attr"`
	Name     string    `xml:"name"`
	Version  string    `xml:"version"`
	PURL     string    `xml:"purl"`
	Licences []License `xml:"licenses>license"`
}

type License struct {
	ID   *string `xml:"id,omitempty"`
	Name *string `xml:"name,omitempty"`
	URL  *string `xml:"url,omitempty"`
}

func NewComponent(m *module.Module) Component {
	mod := Module{
		Path:    m.Path,
		Version: m.Version,
	}
	return Component{
		Type:    "library",
		Name:    m.Path,
		Version: m.Version,
		PURL:    mod.PURL(),
	}
}

func (c *Component) WithLicense(l *license.License, config *config.Config) {
	if l != nil && l.SPDX != "" {
		var url *string

		if u, ok := config.SBOMLicenseURLs[c.Name]; ok {
			url = &u
		}

		c.Licences = []License{
			{
				ID:  &l.SPDX,
				URL: url,
			},
		}
	} else {
		c.WithLicenseFallback(config)
	}
}

func (c *Component) WithLicenseFallback(config *config.Config) {
	if l, ok := config.Override[c.Name]; ok {
		var url *string

		if u, ok := config.SBOMLicenseURLs[c.Name]; ok {
			url = &u
		}

		c.Licences = []License{
			{
				Name: &l,
				URL:  url,
			},
		}
	}
}

type Module struct {
	Path    string `json:"Path"`
	Version string `json:"Version"`
}

func (m Module) PURL() string {
	var ns, n string
	n = m.Path
	chunks := strings.Split(m.Path, "/")

	if len(chunks) > 1 {
		ns = strings.Join(chunks[:len(chunks)-1], "/")
		n = chunks[len(chunks)-1]
	}

	p := packageurl.NewPackageURL(packageurl.TypeGolang, ns, n, m.NormalizeVersion(m.Version), nil, "")
	return p.ToString()
}

func (m Module) NormalizeVersion(v string) string {
	return strings.TrimPrefix(v, "v")
}

// SBOMOutput writes the results of license lookups to an xml file.
type SBOMOutput struct {
	// Path is the path to the file to write. This will be overwritten if
	// it exists.
	Path string

	// Config is the configuration (if any). This will be used to check
	// if a license is allowed or not.
	Config *config.Config

	modules map[*module.Module]interface{}
	lock    sync.Mutex
}

// Start implements Output
func (o *SBOMOutput) Start(m *module.Module) {}

// Update implements Output
func (o *SBOMOutput) Update(m *module.Module, t license.StatusType, msg string) {}

// Finish implements Output
func (o *SBOMOutput) Finish(m *module.Module, l *license.License, err error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if o.modules == nil {
		o.modules = make(map[*module.Module]interface{})
	}

	o.modules[m] = l
	if err != nil {
		o.modules[m] = err
	}
}

// Close implements Output
func (o *SBOMOutput) Close() error {
	o.lock.Lock()
	defer o.lock.Unlock()

	f, err := os.Create(o.Path)
	if err != nil {
		return errors.Wrapf(err, "failed to create file: %s", o.Path)
	}
	defer f.Close()

	bom := BOM{XMLNs: "http://cyclonedx.org/schema/bom/1.1", Version: 1}
	bom.SerialNumber = uuid.New().URN()
	bom.Components = []Component{}

	keys := make([]string, 0, len(o.modules))
	index := map[string]*module.Module{}
	licenses := map[string]interface{}{}

	for m, l := range o.modules {
		keys = append(keys, m.Path)
		index[m.Path] = m
		licenses[m.Path] = l
	}
	sort.Strings(keys)

	for _, k := range keys {
		m := index[k]
		l := licenses[k]
		switch t := l.(type) {
		case error:
			c := NewComponent(m)
			c.WithLicenseFallback(o.Config)
			bom.Components = append(bom.Components, c)
		case *license.License:
			c := NewComponent(m)
			c.WithLicense(t, o.Config)
			bom.Components = append(bom.Components, c)
		default:
			return fmt.Errorf("unexpected license type: %T", t)
		}
	}

	xmlOut, err := xml.MarshalIndent(bom, " ", "  ")
	if err != nil {
		return errors.Wrapf(err, "failed to generate xml")
	}

	_, err = f.Write(xmlOut)
	if err != nil {
		return errors.Wrapf(err, "failed to write xml to a file")
	}

	return nil
}
