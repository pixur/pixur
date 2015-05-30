package pixur

import (
	"database/sql"
)

var (
	rawSchemaTable = []string{
		"CREATE TABLE IF NOT EXISTS `pics` (" +
			"  `id` int(11) NOT NULL AUTO_INCREMENT," +
			"  `data` blob NOT NULL," +
			"  `created_time` bigint(20) NOT NULL," +
			"  `sha256_hash` tinyblob NOT NULL," +
			"  `hidden` bool NOT NULL," +
			"  PRIMARY KEY (`id`)," +
			"  UNIQUE KEY (`sha256_hash`(255))," +
			"  KEY `created_time` (`created_time`)," +
			"  KEY `hidden` (`hidden`)" +
			") ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin AUTO_INCREMENT=1;",

		"CREATE TABLE IF NOT EXISTS `tags` (" +
			"  `id` int(11) NOT NULL AUTO_INCREMENT," +
			"  `data` blob NOT NULL," +
			"  `name` varchar(255) NOT NULL," +
			"  PRIMARY KEY (`id`)," +
			"  UNIQUE KEY (`name`) " +
			"  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin AUTO_INCREMENT=1;",

		"CREATE TABLE IF NOT EXISTS `pictags` (" +
			"  `pic_id` int(11) NOT NULL," +
			"  `tag_id` int(11) NOT NULL," +
			"  `data` blob NOT NULL," +
			"  PRIMARY KEY (`pic_id`,`tag_id`)" +
			"  ) ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_bin;",
	}
)

func createTables(db *sql.DB) error {
	for _, schemaTable := range rawSchemaTable {
		if _, err := db.Exec(schemaTable); err != nil {
			return err
		}
	}
	return nil
}
