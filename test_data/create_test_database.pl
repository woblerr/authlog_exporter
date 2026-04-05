#!/usr/bin/env perl

# Based on https://github.com/maxmind/MaxMind-DB-Writer-perl
use strict;
use warnings;
use utf8;
use MaxMind::DB::Writer::Tree;
use Net::Works::Network;

my $file = '/build/geolite2_test.mmdb';

my %types = (
  accuracy_radius    => 'uint32',
  city               => 'map',
  code               => 'utf8_string',
  continent          => 'map',
  country            => 'map',
  en                 => 'utf8_string',
  ru                 => 'utf8_string',
  geoname_id         => 'uint32',
  iso_code           => 'utf8_string',
  latitude           => 'double',
  location           => 'map',
  longitude          => 'double',
  names              => 'map',
  postal             => 'map',
  time_zone          => 'utf8_string',
  registered_country => 'map',
);

my $tree = MaxMind::DB::Writer::Tree->new(
  ip_version               => 4,
  record_size              => 24,
  database_type            => 'GeoIP2-City',
  languages                => ['en', 'ru'],
  description              => { en => 'Test database', ru => 'Тестовая база данных' },
  map_key_type_callback    => sub { $types{ $_[0] } },
  remove_reserved_networks => 0,
);

# See https://github.com/maxmind/MaxMind-DB/blob/main/source-data/GeoIP2-City-Test.json
my %ips = (
  '12.123.12.123/24' => {
    city => {
      geoname_id => 2618424,
      names      => {
        en => '',
        ru => '',
      },
    },
    country => {
      geoname_id => 6252001,
      iso_code   => 'US',
      names      => {
        en => 'United States',
        ru => 'Соединённые Штаты',
      },
    },
    location => {
      accuracy_radius => 50,
      latitude        => 37.751,
      longitude       => -97.822,
      time_zone       => 'America/Chicago',
    },
    continent => {
      code       => 'NA',
      geoname_id => 6255149,
      names      => {
        en => 'North America',
        ru => 'Северная Америка',
      },
    },
    postal => {
      code => ''
    },
    registered_country => {
      geoname_id => 6252001,
      iso_code => 'US',
      names => {
        en => 'United States',
        ru => 'Соединённые Штаты',
      },
    },
  },
  '123.123.123.0/24' => {
    city => {
      geoname_id => 1816670,
      names      => {
        en => 'Beijing',
        ru => 'Пекин',
      },
    },
    country => {
      geoname_id => 1814991,
      iso_code   => 'CN',
      names      => {
        en => 'China',
        ru => 'Китай',
      },
    },
    location => {
      accuracy_radius => 50,
      latitude        => 39.9042,
      longitude       => 116.4074,
      time_zone       => 'Asia/Shanghai',
    },
    continent => {
      code       => 'AS',
      geoname_id => 6255147,
      names      => {
        en => 'Asia',
        ru => 'Азия',
      },
    },
    postal => {
      code => ''
    },
    registered_country => {
      geoname_id => 1814991,
      iso_code => 'CN',
      names => {
        en => 'China',
        ru => 'Китай',
      },
    },
  },
  '12.123.123.0/24' => {
    city => {
      geoname_id => 2950159,
      names      => {
        en => 'Berlin',
        ru => 'Берлин',
      },
    },
    country => {
      geoname_id => 2921044,
      iso_code   => 'DE',
      names      => {
        en => 'Germany',
        ru => 'Германия',
      },
    },
    location => {
      accuracy_radius => 50,
      latitude        => 52.5200,
      longitude       => 13.4050,
      time_zone       => 'Europe/Berlin',
    },
    continent => {
      code       => 'EU',
      geoname_id => 6255148,
      names      => {
        en => 'Europe',
        ru => 'Европа',
      },
    },
    postal => {
      code => ''
    },
    registered_country => {
      geoname_id => 2921044,
      iso_code => 'DE',
      names => {
        en => 'Germany',
        ru => 'Германия',
      },
    },
  },
  '123.123.12.0/24' => {
    city => {
      geoname_id => 2643743,
      names      => {
        en => 'London',
        ru => 'Лондон',
      },
    },
    country => {
      geoname_id => 2635167,
      iso_code   => 'GB',
      names      => {
        en => 'United Kingdom',
        ru => 'Великобритания',
      },
    },
    location => {
      accuracy_radius => 50,
      latitude        => 51.5074,
      longitude       => -0.1278,
      time_zone       => 'Europe/London',
    },
    continent => {
      code       => 'EU',
      geoname_id => 6255148,
      names      => {
        en => 'Europe',
        ru => 'Европа',
      },
    },
    postal => {
      code => ''
    },
    registered_country => {
      geoname_id => 2635167,
      iso_code => 'GB',
      names => {
        en => 'United Kingdom',
        ru => 'Великобритания',
      },
    },
  },
);

for my $address (keys %ips) {
  my $network = Net::Works::Network->new_from_string(string => $address);

  $tree->insert_network($network, $ips{$address});
}

open my $fh, '>:raw', $file;
$tree->write_tree($fh);
close $fh;
