Changes
===========

**Master (not released)**

  - improvement: make udp msg max size configurable

**0.2 (2017-05.09)**

  - bugfix: use correct flag to open log file
  - bugfix: fix a typo in internal stat metric name
  - improvement: add "Prefix" option to ceilometer receiver

**0.1 (2017-05-09)**

  - implement ceilometer receiver, which receives and converts message sent by [ceilometer udp publisher](https://github.com/openstack/ceilometer/blob/master/ceilometer/publisher/udp.py)
