# BRK3-034: Securing Workplace Safety Detection System

This project is intended for demonstration purposes only. It is not
intended for use in a production environment.

This repository contains a simple demo application showing how to combine Google Cloud services to build a service to detect personal protection equipment (PPE) such as hard hats, within a compliance-constrained project.  

The demo project uses CMEK to protect data-at-rest, GKE Shielded Nodes and Confidential GKE Nodes to protect data while being processed, and Workload Identity to control access to Vertex AI, which hosts Gemini 2.5 Flash for image processing.  Data Boundaries via Assured Workloads (for IL5) provides foundational controls for a regulated environment.


## Getting started
This is meant to be used with the codelab for BRK3-034 and not standalone.

1. Start the codelab
1. Follow the steps in the lab to do the following:
    1. Create the infrastructure within a Data Boundary.
    1. Install the workplace safety detection system demo application on a Confidential GKE cluster.


This is not an officially supported Google product. This project is not
eligible for the [Google Open Source Software Vulnerability Rewards
Program](https://bughunters.google.com/open-source-security).
