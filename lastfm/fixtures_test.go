package lastfm

// Shared XML fixtures used across multiple test files.

const topTagsXML = `<lfm status="ok">
  <toptags>
    <tag><name>heavy metal</name><count>100</count></tag>
    <tag><name>metal</name><count>80</count></tag>
  </toptags>
</lfm>`

const userTagsXML = `<lfm status="ok">
  <tags>
    <tag><name>classic</name></tag>
    <tag><name>nwobhm</name></tag>
  </tags>
</lfm>`

const topArtistsXML = `<lfm status="ok">
  <topartists>
    <artist>
      <name>Iron Maiden</name>
      <playcount>1000</playcount>
    </artist>
    <artist>
      <name>Metallica</name>
      <playcount>800</playcount>
    </artist>
  </topartists>
</lfm>`

const topTracksXML = `<lfm status="ok">
  <toptracks>
    <track>
      <name>The Trooper</name>
      <playcount>500</playcount>
      <artist><name>Iron Maiden</name></artist>
    </track>
  </toptracks>
</lfm>`

const topAlbumsXML = `<lfm status="ok">
  <topalbums>
    <album>
      <name>Piece of Mind</name>
      <playcount>300</playcount>
      <artist><name>Iron Maiden</name></artist>
    </album>
  </topalbums>
</lfm>`

const tagInfoXML = `<lfm status="ok">
  <tag>
    <name>heavy metal</name>
    <reach>12345</reach>
    <taggings>67890</taggings>
    <url>https://www.last.fm/tag/heavy+metal</url>
    <wiki>
      <summary>Heavy metal summary.</summary>
      <content>Heavy metal full content.</content>
    </wiki>
  </tag>
</lfm>`

const lovedTracksXML = `<lfm status="ok">
  <lovedtracks user="testuser">
    <track>
      <name>The Trooper</name>
      <artist><name>Iron Maiden</name></artist>
      <date uts="1609459200">01 Jan 2021</date>
    </track>
  </lovedtracks>
</lfm>`

const weeklyChartListXML = `<lfm status="ok">
  <weeklychartlist user="testuser">
    <chart from="1609459200" to="1610064000"/>
    <chart from="1610064000" to="1610668800"/>
  </weeklychartlist>
</lfm>`

const similarTracksXML = `<lfm status="ok">
  <similartracks track="The Trooper" artist="Iron Maiden">
    <track>
      <name>Run to the Hills</name>
      <match>0.8</match>
      <artist><name>Iron Maiden</name></artist>
    </track>
  </similartracks>
</lfm>`

const noWikiAlbumInfoXML = `<lfm status="ok">
  <album>
    <name>Dance of Death</name>
    <artist>Iron Maiden</artist>
  </album>
</lfm>`

const noWikiTrackInfoXML = `<lfm status="ok">
  <track>
    <name>The Trooper</name>
    <artist><name>Iron Maiden</name></artist>
  </track>
</lfm>`
