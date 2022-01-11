package main

type Site struct {
    SiteName        string   `json:"siteName"`
    RootDir         string   `json:"rootDir"`
    CollectionTypes []string `json:"collectionTypes"`
    SingleTypes     []string `json:"singleTypes"`
}

type SiteService interface {
    CreateEntry()
    RemoveEntry()
    RenameEntry()
    ModifyEntry()
    RebuildSite()
}
