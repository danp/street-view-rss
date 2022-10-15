# street-view-rss

Generate RSS feeds of Street View updates.

Try it here: https://street-view-rss.danp.net/

Inspired by [this tweet](https://twitter.com/mendel/status/1478111813464375296)
and [datasette-atom](https://datasette.io/plugins/datasette-atom).

## How it works

The Google Maps API has a
[Street View Image Metadata](https://developers.google.com/maps/documentation/streetview/metadata)
endpoint which returns info about Street View for a particular location.

This app takes URLs like `/atom.xml?l=<address 1>&l=<address 2>&...`,
extracts the addresses, and checks the metadata endpoint for each one.

It then uses the info returned to generate an RSS feed, setting
entries' updated times based on when Street View was last updated for
each address.

The index page (like [here](https://street-view-rss.danp.net/)) helps
generate links with addresess embedded. The generated link can then be
added to an RSS reader.
