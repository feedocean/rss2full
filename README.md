RSS2Full
====

RSS2Full is a free and cross-platform full text RSS hosting service. Automatic transform summary-only RSS feeds into full-text RSS feeds. Read articles in full, in peace, in your favourite news reading application.

Notes: current v0.1 allows max items of each feed is 10.

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

Start the server in a custom port:

```
rss2full -p 9000
```

## API

```
/feed/<RSS feed url begin with http://>
```

RSS feeds for test full-text:

- https://www.engadget.com/rss.xml

## Installation

```
go get github.com/feedocean/rss2full
```

### Binary

[Download for Windows 64-bit & Linux](https://github.com/feedocean/rss2full/releases)

### Docker

See [Dockerfile](https://github.com/feedocean/rss2full/blob/master/Dockerfile) for image details.

```
docker run -d -p 8088:8088 --name rss2full feedocean/rss2full:latest
```

You can see all the Docker tags [here](https://hub.docker.com/r/feedocean/rss2full/tags).

## Publish on Web

When you're ready to deploy your full-text RSS service to a cloud server, anyone or RSS reader apps can access them.

```
        |==============|
        |   RSS2Full   |
        |==============|
           |       |   
          /         \
         /           \
        /             \
 /-----------\   /-----------\
 |   Feedly  |   |  Inoreader | (RSS apps)
 \-----------/   \-----------/
```

- [Vultr](https://www.vultr.com/?ref=7961474-4F)

New user to get $50 free credit to trial, valid for 1 month.

- [DigitalOcean](https://m.do.co/c/26c25781d4a3)

New user to get $50 free credit to trial

## Usage

Open a web-browser, visit `http://127.0.0.1:8080/`(replacing with your IP address)

![Home](https://user-images.githubusercontent.com/5097328/66846331-430b4d80-efa4-11e9-93d6-f2a0cea1ec64.png)

![engadget](https://user-images.githubusercontent.com/5097328/66851551-8d44fc80-efad-11e9-8ca6-36bca8d5d3cf.png)



