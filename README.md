# Midas

[![Test](https://github.com/kovansky/midas/actions/workflows/test.yml/badge.svg)](https://github.com/kovansky/midas/actions/workflows/test.yml)

## What am I?

Midas is an app that allows you to connect the headless CMS with some static site generator.

## How does it work?

Midas is listening for the webhooks from the "provider" - CMS - and then, depending on the payload, it modifies the
site (e.g. adds a new post) and regenerates it.

## Supported technologies

### CMS-side (providers)

- [Strapi](https://strapi.io/)

### Site generators (receivers)

- [Hugo](https://gohugo.io)

### Support matrix

|      | Strapi |
|------|--------|
| Hugo |    ✔   |

## Installation

### Pre-built binaries

You can find downloads in the [releases section on GitHub](https://github.com/kovansky/midas/releases).

### Install using go

You may want to install the app using go itself. To do that, type following command:

```shell
go install github.com/kovansky/midas/cmd/midasd@latest
```

## Quickstart

### Midas configuration

You need to create a configuration file (can be in your home directory). Here you can find a sample configuration:

```json5
{
  "addr": "127.0.0.1:8445",
  // This is a address that the app will be listening on
  "rollbarToken": "",
  // You can paste the rollbar token here to receive internal errors reported there (https://rollbar.com/)
  "sites": {
    // This is probably most important part of the config - here you specify where your static site code is
    "abcd-efgh-ijkl": {
      // We start with an API key, a.k.a. identifier of the site. In future the codes will be held in some database, not there
      "siteName": "Sample site",
      // Name of the site. May be passed to generator.
      "service": "hugo",
      // Very important setting, specifies which SSG (receiver) is used. Required.
      "rootDir": "/home/kitten/hugo-site",
      // Where the site code lives. Should be absolute path. Required.
      "buildDrafts": false,
      // You can enable to build a site with draft posts along with the main site (in the separate dir). Default: false.
      "draftsUrl": "http://preview.hugo.local",
      // If you enable the option above ^, here you need to pass the URL at which the site will be available, so the generator can build URLs properly.
      "outputSettings": {
        // Here you can set where the static site will be generated (can be absolute or relative - then will be placed under rootDir).
        "build": "public",
        // Main site will be generated to this directory. Default: public
        "draft": "publicDrafts"
        // Site with drafts will be generated to this directory. Default: publicDrafts
      },
      "registry": {
        // Required. Midas keeps an id->filename mapping for created entries.
        "type": "jsonfile",
        // Currently only jsonfile storage is supported.
        "location": "./midas-registry.json"
        // Provide json filename where the mapping should be saved. Can be absolute or relative - then will be placed under site's rootDir 
      },
      "collectionTypes": {
        // List incoming types that should be treated as collections (multiple entries per type).
        "post": {
          // ...here we are allowing "post" type as a collection type, because we will have many posts
          "archetypePath": "archetypes/default.md",
          // We can choose the archetype used to generate content for this type.
          "outputDir": "content/posts/"
          // And specify the directory to which the entries will be saved.
        }
      },
      "singleTypes": {
        // Same as above, but with single types (so type=one entry).
        "homepage": {}
      }
    }
    // Note, that types not listed in collectionTypes nor singleTypes will be ignored.
  }
}
```

### Start midas

After creating the configuration file you need to start the Midas. The main command is `midasd`. It takes two
arguments (optional):

- `config` - path to your config file. Default: `config.json` (in current directory).
- `env` - development or production. Used in rollbar logging. Default: production. Can also be set using MIDAS_ENV
  environmental variable.

So sample startup command could be:

```shell
midasd --config ~/midas.json --env development
```

### Point Strapi to Midas

You need to head to the Settings in Strapi, then to Webhooks, and click "Create new webhook"
(or just
visit [https://your-strapi-installation.com/admin/settings/webhooks/create](http://localhost:1337/admin/settings/webhooks/create))
.

You can choose whatever name you want. In the URL you need to pass an unique URL for the connection. It will have the
following form: `https://midas-installation.com/{{provider}}/{{receiver}}`. For example if you are connecting **Strapi**
and **Hugo**, the URL will be `https://midas-installation.com/strapi/hugo`.

Now go to the Headers and add the `Authorization` header with value `Bearer {{insert API key}}`,
i.e. `Bearer abcd-efgh-ijkl`.

In the Events part it is recommended to select all checkboxes for **Entry**. **Media** is currently not supported.

Whole Strapi settings should look like this:
![strapi-webhook-config.png](images/strapi-webhook-config.png)

### Creating archetypes

When creating archetypes for entries, you can use data from the Payload sent by the Provider. Most of the information
will be stored in two maps: one named `Entry` (with entry data, like values of the fields from CMS), and second
named `Metadata` with some generated data, like information if entry is published. Sample archetype for Strapi->Hugo
relation may look like this:

```html
---
title: "{{ index .Entry "Title" }}" # We read the title from one of the fields configured in CMS
date: {{ index .Metadata "createdAt" }} # We read the publication date from metadata
draft: {{ not (index .Metadata "published") }} # As well as the information if the post is published or not.
---

<h3>{{ index .Entry "Subtitle" }}</h3> <!-- We may for example include some subtitle -->
<div id="entry-content">
    {{ index .Entry "Content" }} <!-- And after all we include the main entry content -->
</div>
```

## Feature requests? Bugs?

You are welcome to open an issue.

## Contributing?

You are welcome to write new features, providers or receivers :) 

## License

Project is released under GNU GPLv3. For more information, see LICENSE file.

## Author

Created and maintained by [F4 Developer (Stanisław Kowański)](https://www.f4dev.me)
