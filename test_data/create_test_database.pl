#!/usr/bin/env perl

# Based on https://github.com/maxmind/MaxMind-DB-Writer-perl
use strict;
use warnings;
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
  languages                => ['en'],
  description              => { en => 'Test database' },
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
      },
    },
    country => {
      geoname_id => 6252001,
      iso_code   => 'US',
      names      => {
        en => 'United States',
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
