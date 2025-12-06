package contentfulmanagement

func NewTaxonomyConceptLink(id string) TaxonomyConceptLink {
	return TaxonomyConceptLink{
		Sys: NewTaxonomyConceptLinkSys(id),
	}
}

func NewTaxonomyConceptLinkSys(id string) TaxonomyConceptLinkSys {
	return TaxonomyConceptLinkSys{
		Type:     TaxonomyConceptLinkSysTypeLink,
		LinkType: TaxonomyConceptLinkSysLinkTypeTaxonomyConcept,
		ID:       id,
	}
}
