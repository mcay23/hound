package view

type ProvidersResponseObject struct {
	MediaSource		  string						`json:"media_source"`
	SourceID		  int    						`json:"source_id"`
	MediaType		  string                    	`json:"media_type"`  	   // movies or tvshows, etc.
	IMDbID            string                    	`json:"imdb_id"`           // starts with 'tt'
	Season            int                       	`json:"season,omitempty"`  // shows only
	Episode           int                       	`json:"episode,omitempty"`
	Providers 		  *[]ProviderObject				`json:"providers"`
}

type ProviderObject struct {
	Provider          string                    	`json:"provider"`	 	   // provider name in /providers folder
	Streams           *[]StreamObject			   	`json:"streams"`
}

type StreamObject struct {
	Addon			 string							`json:"addon"`
	Cached 			 string							`json:"cached"`    // whether the stream is cached
	Service          string							`json:"service"`   // service such as RD, etc.
	P2P				 string 						`json:"p2p"`	   // type of stream, such as 'debrid'
	InfoHash		 string							`json:"infohash"`
	Indexer          string							`json:"indexer"`
	Filename		 string							`json:"file_name"`
	FolderName		 string							`json:"folder_name"` // folder name for packs
	Resolution	     string							`json:"resolution"`
	FileIndex		 int							`json:"file_idx"`  // file index of stream in torrent
	FileSize		 int							`json:"file_size"` // file size in bytes
	Rank			 int							`json:"rank"`
	Seeders			 int							`json:"seeders"`
	Leechers		 int 							`json:"leechers"`
	URL              string							`json:"url"`
	ParsedData		 *ParsedData					`json:"data"`
}

type ParsedData struct {
	VideoCodec		 string							`json:"codec"`
	AudioCodec	     []string					 	`json:"audio"`
	AudioChannels    []string						`json:"channels"`
	FileContainer    string							`json:"container"`
	Languages		 []string						`json:"languages"`
	BitDepth		 string  						`json:"bit_depth"`        // eg. 10bit
	HDR				 []string						`json:"hdr"`
}
