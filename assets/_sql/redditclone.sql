SET NAMES utf8;
SET time_zone = '+00:00';
SET foreign_key_checks = 0;
SET sql_mode = 'NO_AUTO_VALUE_ON_ZERO';

DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `username` varchar(100),
  `password` varchar(100),
  `user_id` int(11) AUTO_INCREMENT,
  PRIMARY KEY (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

DROP TABLE IF EXISTS `authorization`;
CREATE TABLE `authorization` (
  `token` varchar(255),
  `time_to` int(11)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
