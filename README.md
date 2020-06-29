__go-awx-approver__
<img align="right" src="logo/logo.png" alt="logo" width="300">

Web application for implementing a request &amp; approval process for running Ansible Tower (AWX) templates.


Deps:  
- github.com/joncalhoun/form
- github.com/Colstuwjx/awx-go
- github.com/Sirupsen/logrus
- github.com/kelseyhightower/envconfig
- github.com/google/uuid


# To do
- Don't render the data with go:
  - Render only the plain template without data
  - Provide data like cart items or requests by API
  - Use jQuery to get the data from an API and populate the HTML

# Notifications
- Email (only when set in config file, otherwise disabled)
- Slack " (write to channel or DM)
