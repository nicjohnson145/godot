package lib

type RepoConfig struct {
	Files          StringMap                `json:"files"`
	Bootstraps     map[string]Bootstrap     `json:"bootstraps"`
	GithubReleases map[string]GithubReleaseConfiguration `json:"github_releases"`
	Hosts          map[string]Host          `json:"hosts"`
}

func (c *RepoConfig) makeMaps() {
	if c.Files == nil {
		c.Files = make(StringMap)
	}

	if c.Bootstraps == nil {
		c.Bootstraps = make(map[string]Bootstrap)
	}

	if c.Hosts == nil {
		c.Hosts = make(map[string]Host)
	}

	if c.GithubReleases == nil {
		c.GithubReleases = make(map[string]GithubReleaseConfiguration)
	}
}

func (c *RepoConfig) setGithubReleaseNames() {
	for name := range c.GithubReleases {
		current := c.GithubReleases[name]
		current.name = name
		c.GithubReleases[name] = current
	}
}
