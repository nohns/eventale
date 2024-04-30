# Eventale - a magical Go-based event store 🪄✨

Tell all events to Eventale, and let it recite to you the entire story (fairytale, if you will) so far.
Or just ask it to recite to you the entire story of a given subject. You can then yourself piece together
a picture of what that subject might look like right now, a week ago or maybe even years ago. With Eventale,
you got the entire history of changes to your data, so you don't loose valuable data when the state of
your application changes. And when it changes, you got the possibility of drawing your own read models
optimized for specific queries, through consuming the events published by Eventale.

**NOTE: in development**

## Purpose

This is a hobby project, to go more in depth with self-rolled TCP protocol, protoc plugins, event distribution
and more. Built entirely in Go.

## Components

Eventale consists of serveral components:

- An `taled` event store server, built from `cmd/taled`
- A Go library in the form of the root of this repository
- The `alice` CLI application for interacting with the server
- A `protoc` plugin for generating Eventale ready Go protobuf structs
