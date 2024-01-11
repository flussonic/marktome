---
keywords: Live streaming
description: Flussonic Media Server can retransmit streaming video 
---

# Live Streaming {#live-page}

*Flussonic Media Server* can retransmit streaming video into multiple output formats on the fly with just-in-time packaging. Use [Flussonic API reference](<m>flussonic-api</m>#tag/stream/operation/stream_save%7Cbody%7Cdrm) for DRM.

*Flussonic Media Server* supports three types of streams:

- `static` — streams that are being broadcasted all the time.

- `live` — user-published streams. See <link anchor="live-publish-page">Publishing</link> for details.

**Contents:**

* <link anchor="live-static">Static streams</link>
* <link anchor="live-ondemand">On-demand streams</link>
* <link anchor="live-playback_urls">Stream playback</link>


!!! caution
    <m>spec_chars_note_ru</m>


## Static streams {#live-static}

Static streams are launched upon a start of the server. Flussonic continuously monitors static streams.

<!--
This is a comment.

Should not get out of MD
-->

This is a copy if static.conf snippet:

<include-snippet id="static.conf" />

The format of a stream definition in the **/etc/flussonic/flussonic.conf** file is:

<snippet id="static.conf">
stream example_stream {
  input udp://239.0.0.1:1234;

}
</snippet>

In this example:

* `example` is the name that must be used to request the stream from Flussonic Media Server.
* `udp://239.0.0.1:1234` is the data source URL.

**Important.** The name of a stream should contain only Latin characters, digits, dots (`.`), dashes (`-`), and underscores (`_`).
If the name contains any other characters, DVR and live streams might work incorrectly.

**To add a stream via the web interface:**

Go to the **Media** tab and click **Add** next to **Streams**.

![Flussonic add stream](<m>img_dir</m>/auto/live/static.js/media.png)

Then enter the name of the stream and the data source URL. Click **Create** and Flussonic will add the stream to the list.

To specify this kind of behavior, change the stream type to `ondemand`:

<snippet id="ondemand1.conf">
ondemand ipcam {
  input rtsp://localhost:554/source;
}
</snippet>

## drmnow! DRM {#drm-drmnow-page}

Now example of CAS server

`cas-server.php`:

```php
<?php
  header("HTTP/1.0 200 OK");
  $resource = $_GET["file"];
  $number = $_GET["number"];
  
  header("Content-Length: ".strlen($key));
  echo $key;
?>
```


![Flussonic ondemand](<m>img_dir</m>/auto/live/ondemand.js/overview.png)

!!! caution
    If Media Server ingests the `ondemand` source stream using RTMP, RTSP, or HTTP MPEG-TS protocols, there will be some complications with outputting HLS streams. This is because those streaming protocols require 10-30 second buffering.

You can specify the stream's lifetime after a client has disconnected:

