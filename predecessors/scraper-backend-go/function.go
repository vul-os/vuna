package function

import (
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/exolutionza/scraper-backend-go/internal/router"

)

func init() {
	functions.HTTP("Router", router.Router)
}

