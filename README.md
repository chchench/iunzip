# iUnzip

Just some code written for PoC and benchmarking that others might also find useful


## Background

No, this is not about about a better way to unzip file, as the name might suggest, but rather, it's about a better way to process the content within an archive file over the Web--which is useful in certain domains like cybersecurity where detection of potential threats within files, including archive files, are often done.

When processing of a larger file is done locally, the most expensive part of examining its content is often the computational processing along with the I/O operations required to retrieve the data to be processed. However, when the processing of the file needs to be done remotely from the original location of the content, the latency required for the file to be transferred across the network before it can be processed can dominate and contribute significantly to the overall turnaround time--even with abundance of computational resources.


## Goal for Proof of Concept

The code published in this repo was written as part of a PoC to confirm using real-world data that for certain applications where the files are of certain variety and mix, that decompressing of much large (albeit relatively few) archive files into much smaller individual files, and transferring them concurrenty over to a remote cloud-based service that offers high-performance, scalable processing of files, overall processing time can be reduced.


## Design Requirements

Decompression of archive (and subsequent processing) should be controlled based on available host resource so as to not starve user's application. The ceiling can be determined by available storage, CPU capacity, available memory, etc.

Decompressing large archives in one shot may unnecessarily block the client for too long a duration. For smoother scheduling, it might be better to decompose the archive file in phases or in small steps, while incurring some additional "disk" I/O operations. Given more computers have migrated off the traditional spinning media storage devices, the additional I/O overhead should be negligible compared to network I/O.

In the future, for certain computing environments, it might also be possible to fine-tuning performance by NICE value, cgroup, CPU usage capping, etc.


## PoC Implementation

The program provided does the following:

* Detect whether the target file to scan is a ZIP file or not. If not, just exit.
* If target file is a ZIP file, it will iterate through all the member files.
* If a member file is also a ZIP archive file, it will also recursively decompress it.
* Otherwise, for every (non-archive) member file, it will start a new worker thread to process the file; work is simulated by the thread sleeping for a random period of time. 
* At most a certain number of jobs (and worker threads) are allowed to execute concurrently. Number of executing worker jobs/threads is the only factor determining whether a new job can be started or not, in this published version.


## License

Making the source code to this app available under
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
