RSS2Full
====

RSS2Full is a free and cross-platform full text RSS hosting service. Automatic transform summary-only RSS feeds into full-text RSS feeds. Read articles in full, in peace, in your favourite news reading application.

Notes: current v0.1 allows max items of feed is 10.

## Features 

- Free, fast and reliable.
- Easy to deploy and run on a local or cloud server(web).
- Provides HTTP API interface, easy integrate it into your own RSS service. 

## Command-line usage

```
Usage:
  rss2full  [options...]

Options:
  -a <addr>                 Bind address [default: *]
  -p <port>                 Bind port [default: 8088]
```

## API

```
/feed/<RSS feed url begin with http://>
```

RSS feeds for test full-text:

- https://www.engadget.com/rss.xml

## Installation

### Binary

[Free Download](https://github.com/feedocean/rss2full/releases)

## Usage

Open a web-browser, visit `http://127.0.0.1:8080/`(replacing with your IP address)

![Home](https://user-images.githubusercontent.com/5097328/66846331-430b4d80-efa4-11e9-93d6-f2a0cea1ec64.png)

![engadget](https://user-images.githubusercontent.com/5097328/66851551-8d44fc80-efad-11e9-8ca6-36bca8d5d3cf.png)



