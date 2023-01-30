Name
====

OpenResty - Turning Nginx into a Full-Fledged Scriptable Web Platform

Table of Contents
=================

* [Name](#name)
* [Description](#description)
    * [For Users](#for-users)
    * [For Bundle Maintainers](#for-bundle-maintainers)
* [Additional Features](#additional-features)
    * [resolv.conf parsing](#resolvconf-parsing)
* [Mailing List](#mailing-list)
* [Report Bugs](#report-bugs)
* [Copyright & License](#copyright--license)

Description
===========

OpenResty is a full-fledged web application server by bundling the standard nginx core,
lots of 3rd-party nginx modules, as well as most of their external dependencies.

This bundle is maintained by Yichun Zhang (agentzh).

Because most of the nginx modules are developed by the bundle maintainers, it can ensure
that all these modules are played well together.

The bundled software components are copyrighted by the respective copyright holders.

The homepage for this project is on [openresty.org](https://openresty.org/).

For Users
---------

Visit the [download page](https://openresty.org/en/download.html) on the `openresty.org` web site
to download the latest bundle tarball, and
follow the installation instructions in the [installation page](https://openresty.org/en/installation.html).

For Bundle Maintainers
----------------------

The bundle's source is at the following git repository:

https://github.com/openresty/openresty

To reproduce the bundle tarball, just do

```
make
```

at the top of the bundle source tree.

Please note that you may need to install some extra dependencies, like `perl`, `dos2unix`, and `mercurial`.
On Fedora 22, for example, installing the dependencies
is as simple as running the following commands:

```
sudo dnf install perl dos2unix mercurial
```

[Back to TOC](#table-of-contents)

Additional Features
===================

In additional to the standard nginx core features, this bundle also supports the following:

[Back to TOC](#table-of-contents)

resolv.conf parsing
--------------------

**syntax:** *resolver address ... [valid=time] [ipv6=on|off] [local=on|off|path]*

**default:** *-*

**context:** *http, stream, server, location*

Similar to the [`resolver` directive](https://nginx.org/en/docs/http/ngx_http_core_module.html#resolver)
in standard nginx core with additional support for parsing additional resolvers from the `resolv.conf` file
format.

When `local=on`, the standard path of `/etc/resolv.conf` will be used. You may also specify arbitrary
path to be used for parsing, for example: `local=/tmp/test.conf`.

When `local=off`, parsing will be disabled (this is the default).

This feature is not available on Windows platforms.

[Back to TOC](#table-of-contents)

Mailing List
============

You're very welcome to join the English OpenResty mailing list hosted on Google Groups:

https://groups.google.com/group/openresty-en

The Chinese mailing list is here:

https://groups.google.com/group/openresty

[Back to TOC](#table-of-contents)

Report Bugs
===========

You're very welcome to report issues on GitHub:

https://github.com/openresty/openresty/issues

[Back to TOC](#table-of-contents)

Copyright & License
===================

The bundle itself is licensed under the 2-clause BSD license.

Copyright (c) 2011-2019, Yichun "agentzh" Zhang (章亦春) <agentzh@gmail.com>, OpenResty Inc.

This module is licensed under the terms of the BSD license.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

* Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
* Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS
IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED
TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A
PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED
TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

[Back to TOC](#table-of-contents)

