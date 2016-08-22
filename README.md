#Vesper SQLite

In August 2015, the [Vesper notes app]() [closed
down](http://inessential.com/2016/08/21/last_vesper_update_sync_shutting_down).
The app allows you to export your notes into Dropbox, iCloud Drive, or another
cloud service provider (you can find this option by scrolling to the bottom of
your tags).

The generated notes are text files, but this tool generates a SQLite tool from
them. To do so, run:

```
go get github.com/thomasdenney/vesper-sqlite
vesper-sqlite <PATH TO VESPER FILES>
```

License is Apache 2.
