surfsticker
===========

I wrote `surfsticker` so that I can have a browser window on my desktop that
is reasonably sticky. What I did not want was the default behavior of `surf`, 
where every invocation opens a new window.

Instead, I wrap `surf` in `surfsticker`, with an optional `-sticker` flag.

`surfsticker` creates a `surf` window with an extra XProperty on it named 
`_STICKER`. Without an explicitly chosen `-sticker` flag, it will set the
value to "default".

Once a `surf` window has been started by `surfsticker` with a particular
`_STICKER`, further invocations of `surfsticker` will re-use the existing
window to visit a chosen URL.

Each `_STICKER` is tied to a stylesheet in `~/.surf/styles`.

I have a custom `surf` stylesheet that I like to use with `godoc` so that its
brightness doesn't melt my eyes. To invoke it, I put the following in my 
`.vimrc`:
```
let g:go_play_browser_command = 'surfsticker -sticker godoc %URL% &'
```
This opens a surf window labeled with the  `godoc` `_STICKER` that uses 
the following stylesheet: `~/.surf/styles/godoc.css`. So long as that window 
stays open, it will be reused every time I 
`surfsticker -sticker godoc https://example.com`.

![surfsticker screenshot](https://alrs.tilde.team/surfsticker.png)

installation
============
So long as the `go` toolchain is available and you've added `~/go/bin` to your
regular path, `surfsticker` should install and be usable by invoking
`go install ./...` in its root.

`surfsticker` uses the Linux `inotify` API.
