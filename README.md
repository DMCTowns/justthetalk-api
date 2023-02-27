# JUSTtheTalk - API

This is the API component of the JUSTtheTalk Discussion Forum. It is written using Go/GorillaMux/GORM

## Project Status

The project is incomplete and still under development prior to the first production release however a beta version is running on https://www.justthetalk.co.uk. While mostly feature complete the project requires refactoring and polish.

One of the key issues is the use of panics to pass error state back to the user. It works but it's not idiomatic and generally a bit rubbish. This needs fixing. I'd also factor out database access and other code behind an interface. I often write code this way - the first pass is just to explore the problem domain and perhaps try out a few ideas as quickly as possible, the results can end up being a bit messy. Once I know what I'mm dealing with then subsequent passes tidy and optimise the code.

There are still some design decisions to be finalised. Some objects are currently cached in Redis, I'm not sure there is much benefit to this at the moment. I might rip it out.

Test coverage needs to be improved.
