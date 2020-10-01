__AWX-Judge__
<img align="right" src="logo/logo.png" alt="logo" width="300">

Web application for implementing a request &amp; approval process for running Ansible Tower (AWX) templates.

# About
AWX-Judge is a abstraction layer for RedHat Ansible Tower/AWX, which introduces a request and approval process for running job templates.  
__This project is currently in "work in progress" state.__

# Requirements
AWX-Judge uses MongoDB as a database backend.  

# Features

### Job template import
During the import of job templates, the following changes/additions are possible:
- Add a logo
- Change the name and description
- Attach regular expressions to survey variables

Key feature here are the regular expressions. In Tower/AWX it's not possible to configure regex for survey parameters. You could only verify the inputs during your playbook run, which will result in a failed run if something was not right. With AWX-Judge, all inputs can be verified before a playbook is started.  

### Ordering process
Much like an online shop, the user will get a "cart" containing all his items. After shopping the user can order his items and they will be converted to requests, ready to be reviewed.  
While the items are in the user's cart, he has the ability to:  
- view
- clone/duplicate
- delete

The requests will get locked after the order has been executed and the user can no longer change any parameters.  
After ordering, the user has the following abilities to interact with the requests:  
- view
- reorder (create a clone and add it to the cart)

A reviewer can now check the inputs of the user, as well as if the requests makes sense at all. When the reviewer has come to an decision, he can approve or deny the request. While doing so, he must provide a reason for his decision.  
After approval:
- AWX-Judge will start the job template on tower with the provided variables
- The status of the request will get updated depending on the status of the job in Tower/AWX

After denial:
- The status of the request will get updated and the user will be able to see the reason, why his request got denied
- He can then reorder the request and change hin input parameters

# Planned features
- OIDC for authentication
- Slack bot integration (notify me when my request status changes etc.)
