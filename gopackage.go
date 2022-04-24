package gomodurl

import (
	"log"
	"os"
	"strings"
)

const godocHost = "https://pkg.go.dev"

type GoPackage struct {
	Import         string
	VersionControl string
	Repository     string
	Display        string

	Host       string
	Branch     string
	PathSuffix string
}

func (gp *GoPackage) AddPathSuffix(suffix string) *GoPackage {
	return &GoPackage{
		Import:         gp.Import,
		VersionControl: gp.VersionControl,
		Repository:     gp.Repository,
		Display:        gp.Display,
		Host:           gp.Host,
		Branch:         gp.Branch,
		PathSuffix:     suffix,
	}
}

func (gp *GoPackage) Godoc() string {
	url := []string{
		godocHost,
		gp.Import,
		gp.PathSuffix,
	}
	return strings.Join(url, "/")
}

type GoPackageTree struct {
	next map[rune]*GoPackageTree
	pkg  *GoPackage
}

func NewGoPackageTree() *GoPackageTree {
	return &GoPackageTree{next: map[rune]*GoPackageTree{}}
}

func (gpt *GoPackageTree) Lookup(name string) *GoPackage {
	node := gpt
	cut := 0
	var pkg *GoPackage
	for i, c := range name {
		r := rune(c)
		if node.pkg != nil {
			pkg = node.pkg
			cut = i + 1
		}
		if next := node.next[r]; next != nil {
			node = next
		} else {
			break
		}
	}

	if pkg == nil {
		return nil
	}

	suffix := name[cut:]
	return pkg.AddPathSuffix(suffix)
}

type GoPackageList struct {
	hosts map[string]*GoPackageTree

	hostLength map[string]int
	length     int
}

func NewGoPackageList() *GoPackageList {
	return &GoPackageList{
		hosts:      map[string]*GoPackageTree{},
		hostLength: map[string]int{},
		length:     0,
	}
}

func (gpl *GoPackageList) Lookup(host, path string) *GoPackage {
	if os.Getenv("GOMODURL_ANYHOST") != "" {
		for _, list := range gpl.hosts {
			if pkg := list.Lookup(path); pkg != nil {
				return pkg
			}
		}
	}
	list := gpl.hosts[host]
	if list == nil {
		return nil
	}
	return list.Lookup(path)
}

func (gpl *GoPackageList) Add(pkgs ...*GoPackage) {
	for _, pkg := range pkgs {
		name := strings.TrimPrefix(pkg.Import, pkg.Host)
		name = strings.TrimPrefix(name, "/")
		name = strings.TrimSuffix(name, "/")

		if gpl.hosts[pkg.Host] == nil {
			gpl.hosts[pkg.Host] = NewGoPackageTree()
		}

		node := gpl.hosts[pkg.Host]
		prev := node
		for _, c := range name {
			r := rune(c)
			if node.next[r] == nil {
				node.next[r] = NewGoPackageTree()
			}
			prev = node
			node = node.next[r]
		}
		prev.pkg = pkg

		log.Printf("registered '%s' => '%s'", pkg.Import, pkg.Repository)
		gpl.hostLength[pkg.Host]++
		gpl.length++
	}
}
