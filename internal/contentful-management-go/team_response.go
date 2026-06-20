package contentfulmanagement

func TeamFromGetTeamResponse(response GetTeamRes) (Team, bool) {
	switch response := response.(type) {
	case *GetTeamApplicationJSONOK:
		if response == nil {
			return Team{}, false
		}

		return Team(*response), true
	case *GetTeamApplicationVndContentfulManagementV1JSONOK:
		if response == nil {
			return Team{}, false
		}

		return Team(*response), true
	default:
		return Team{}, false
	}
}
