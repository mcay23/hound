package workers

func InitializeWorkers(downloadWorkers int, ingestWorkers int) {
	InitializeDownloadWorkers(downloadWorkers)
	InitializeIngestWorkers(ingestWorkers)
}
