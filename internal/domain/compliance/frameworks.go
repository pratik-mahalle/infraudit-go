package compliance

// CISAWSControls returns CIS AWS Foundations Benchmark controls
func CISAWSControls() []*Control {
	return []*Control{
		// 1 - Identity and Access Management
		{ControlID: "1.1", Title: "Maintain current contact details", Category: "Identity and Access Management", Severity: SeverityLow, Description: "Ensure contact email and telephone details for AWS accounts are current and map to more than one individual in your organization."},
		{ControlID: "1.2", Title: "Ensure security contact information is registered", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "AWS provides customers with the option of specifying the contact information for account's security team."},
		{ControlID: "1.4", Title: "Ensure no root access keys exist", Category: "Identity and Access Management", Severity: SeverityCritical, Description: "The root user is the most privileged AWS user. AWS Access Keys provide programmatic access to a given AWS account."},
		{ControlID: "1.5", Title: "Ensure MFA is enabled for the root user", Category: "Identity and Access Management", Severity: SeverityCritical, Description: "The root user is the most privileged user in an AWS account."},
		{ControlID: "1.6", Title: "Ensure hardware MFA is enabled for the root user", Category: "Identity and Access Management", Severity: SeverityCritical, Description: "The root user is the most privileged user in an AWS account."},
		{ControlID: "1.7", Title: "Eliminate use of root user for administrative tasks", Category: "Identity and Access Management", Severity: SeverityHigh, Description: "With the creation of an AWS account, a root user is created that cannot be disabled or deleted."},
		{ControlID: "1.8", Title: "Ensure IAM password policy requires minimum length of 14", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "Password policies are used to enforce password complexity requirements."},
		{ControlID: "1.9", Title: "Ensure IAM password policy prevents password reuse", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "IAM password policies can prevent the reuse of a given password by the same user."},
		{ControlID: "1.10", Title: "Ensure MFA is enabled for all IAM users with a console password", Category: "Identity and Access Management", Severity: SeverityHigh, Description: "Multi-Factor Authentication adds an extra layer of protection on top of a username and password."},
		{ControlID: "1.11", Title: "Do not setup access keys during initial user setup for all IAM users", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "AWS console defaults to no check boxes selected when creating a new IAM user."},
		{ControlID: "1.12", Title: "Ensure credentials unused for 45 days or greater are disabled", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "AWS IAM users can access AWS resources using different types of credentials."},
		{ControlID: "1.13", Title: "Ensure there is only one active access key for any single IAM user", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "Access keys are long-term credentials for an IAM user."},
		{ControlID: "1.14", Title: "Ensure access keys are rotated every 90 days or less", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "Access keys consist of an access key ID and secret access key."},
		{ControlID: "1.15", Title: "Ensure IAM users receive permissions only through groups", Category: "Identity and Access Management", Severity: SeverityMedium, Description: "IAM users are granted access to services, functions, and data through IAM policies."},
		{ControlID: "1.16", Title: "Ensure IAM policies with full \"*:*\" administrative privileges are not attached", Category: "Identity and Access Management", Severity: SeverityHigh, Description: "IAM policies are the means by which privileges are granted to users, groups, or roles."},
		{ControlID: "1.17", Title: "Ensure a support role has been created for incident handling", Category: "Identity and Access Management", Severity: SeverityLow, Description: "AWS provides a support center that can be used for incident notification and response."},

		// 2 - Storage
		{ControlID: "2.1.1", Title: "Ensure S3 bucket has server-side encryption enabled", Category: "Storage", Severity: SeverityHigh, Description: "Amazon S3 provides a variety of server-side encryption options to help protect data at rest."},
		{ControlID: "2.1.2", Title: "Ensure S3 bucket policy is set to deny HTTP requests", Category: "Storage", Severity: SeverityMedium, Description: "At the Amazon S3 bucket level, you can configure permissions through a bucket policy."},
		{ControlID: "2.1.3", Title: "Ensure MFA Delete is enabled on S3 buckets", Category: "Storage", Severity: SeverityMedium, Description: "Once MFA Delete is enabled on your S3 versioned bucket it requires additional authentication."},
		{ControlID: "2.1.4", Title: "Ensure all data in Amazon S3 has been discovered and classified", Category: "Storage", Severity: SeverityMedium, Description: "Amazon Macie is a fully managed data security and data privacy service."},
		{ControlID: "2.1.5", Title: "Ensure S3 buckets are configured with Block Public Access", Category: "Storage", Severity: SeverityCritical, Description: "Amazon S3 provides Block Public Access settings for buckets."},
		{ControlID: "2.2.1", Title: "Ensure EBS volume encryption is enabled", Category: "Storage", Severity: SeverityHigh, Description: "Elastic Compute Cloud (EC2) supports encryption at rest when using the Elastic Block Store (EBS)."},
		{ControlID: "2.3.1", Title: "Ensure RDS database instances are encrypted at rest", Category: "Storage", Severity: SeverityHigh, Description: "Amazon RDS encrypted DB instances use the industry standard AES-256 encryption algorithm."},

		// 3 - Logging
		{ControlID: "3.1", Title: "Ensure CloudTrail is enabled in all regions", Category: "Logging", Severity: SeverityHigh, Description: "AWS CloudTrail is a web service that records AWS API calls for your account."},
		{ControlID: "3.2", Title: "Ensure CloudTrail log file validation is enabled", Category: "Logging", Severity: SeverityMedium, Description: "CloudTrail log file validation creates a digitally signed digest file."},
		{ControlID: "3.3", Title: "Ensure the S3 bucket used to store CloudTrail logs is not publicly accessible", Category: "Logging", Severity: SeverityCritical, Description: "CloudTrail logs a record of every API call made in your AWS account."},
		{ControlID: "3.4", Title: "Ensure CloudTrail trails are integrated with CloudWatch Logs", Category: "Logging", Severity: SeverityMedium, Description: "AWS CloudTrail is a web service that records AWS API calls made in a given AWS account."},
		{ControlID: "3.5", Title: "Ensure AWS Config is enabled in all regions", Category: "Logging", Severity: SeverityMedium, Description: "AWS Config is a web service that performs configuration management of supported AWS resources."},
		{ControlID: "3.6", Title: "Ensure S3 bucket access logging is enabled on the CloudTrail S3 bucket", Category: "Logging", Severity: SeverityMedium, Description: "S3 Bucket Access Logging generates a log that contains access records for each request made."},
		{ControlID: "3.7", Title: "Ensure CloudTrail logs are encrypted at rest using KMS CMKs", Category: "Logging", Severity: SeverityMedium, Description: "AWS CloudTrail is a web service that records AWS API calls for an account."},
		{ControlID: "3.8", Title: "Ensure rotation for customer-created CMKs is enabled", Category: "Logging", Severity: SeverityMedium, Description: "AWS Key Management Service (KMS) allows customers to rotate the backing key."},
		{ControlID: "3.9", Title: "Ensure VPC flow logging is enabled in all VPCs", Category: "Logging", Severity: SeverityMedium, Description: "VPC Flow Logs is a feature that enables you to capture information about the IP traffic."},

		// 4 - Monitoring
		{ControlID: "4.1", Title: "Ensure a log metric filter and alarm exist for unauthorized API calls", Category: "Monitoring", Severity: SeverityMedium, Description: "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs."},
		{ControlID: "4.2", Title: "Ensure a log metric filter and alarm exist for Management Console sign-in without MFA", Category: "Monitoring", Severity: SeverityMedium, Description: "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs."},
		{ControlID: "4.3", Title: "Ensure a log metric filter and alarm exist for root account usage", Category: "Monitoring", Severity: SeverityMedium, Description: "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs."},
		{ControlID: "4.4", Title: "Ensure a log metric filter and alarm exist for IAM policy changes", Category: "Monitoring", Severity: SeverityMedium, Description: "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs."},
		{ControlID: "4.5", Title: "Ensure a log metric filter and alarm exist for CloudTrail configuration changes", Category: "Monitoring", Severity: SeverityMedium, Description: "Real-time monitoring of API calls can be achieved by directing CloudTrail Logs to CloudWatch Logs."},

		// 5 - Networking
		{ControlID: "5.1", Title: "Ensure no security groups allow ingress from 0.0.0.0/0 to port 22", Category: "Networking", Severity: SeverityCritical, Description: "Security groups provide stateful filtering of ingress/egress network traffic to AWS resources."},
		{ControlID: "5.2", Title: "Ensure no security groups allow ingress from 0.0.0.0/0 to port 3389", Category: "Networking", Severity: SeverityCritical, Description: "Security groups provide stateful filtering of ingress/egress network traffic to AWS resources."},
		{ControlID: "5.3", Title: "Ensure the default security group of every VPC restricts all traffic", Category: "Networking", Severity: SeverityHigh, Description: "A VPC comes with a default security group whose initial settings deny all inbound traffic."},
		{ControlID: "5.4", Title: "Ensure routing tables for VPC peering are \"least access\"", Category: "Networking", Severity: SeverityMedium, Description: "A VPC peering connection is a networking connection between two VPCs."},
	}
}

// NIST80053Controls returns NIST 800-53 Rev 5 controls
func NIST80053Controls() []*Control {
	return []*Control{
		// Access Control
		{ControlID: "AC-1", Title: "Policy and Procedures", Category: "Access Control", Severity: SeverityLow, Description: "Develop, document, and disseminate access control policy and procedures."},
		{ControlID: "AC-2", Title: "Account Management", Category: "Access Control", Severity: SeverityHigh, Description: "Define and document the types of accounts allowed and specifically prohibited."},
		{ControlID: "AC-3", Title: "Access Enforcement", Category: "Access Control", Severity: SeverityHigh, Description: "Enforce approved authorizations for logical access to information and system resources."},
		{ControlID: "AC-4", Title: "Information Flow Enforcement", Category: "Access Control", Severity: SeverityHigh, Description: "Enforce approved authorizations for controlling the flow of information."},
		{ControlID: "AC-5", Title: "Separation of Duties", Category: "Access Control", Severity: SeverityMedium, Description: "Separate duties of individuals to reduce risk of malevolent activity."},
		{ControlID: "AC-6", Title: "Least Privilege", Category: "Access Control", Severity: SeverityHigh, Description: "Employ the principle of least privilege for specific duties and information systems."},
		{ControlID: "AC-7", Title: "Unsuccessful Logon Attempts", Category: "Access Control", Severity: SeverityMedium, Description: "Enforce a limit of consecutive invalid logon attempts by a user."},
		{ControlID: "AC-11", Title: "Device Lock", Category: "Access Control", Severity: SeverityMedium, Description: "Prevent access to the system by initiating a session lock."},
		{ControlID: "AC-17", Title: "Remote Access", Category: "Access Control", Severity: SeverityHigh, Description: "Establish usage restrictions and implementation guidance for remote access."},
		{ControlID: "AC-18", Title: "Wireless Access", Category: "Access Control", Severity: SeverityHigh, Description: "Establish configuration requirements and usage restrictions for wireless access."},

		// Audit and Accountability
		{ControlID: "AU-1", Title: "Policy and Procedures", Category: "Audit and Accountability", Severity: SeverityLow, Description: "Develop, document, and disseminate audit and accountability policy."},
		{ControlID: "AU-2", Title: "Event Logging", Category: "Audit and Accountability", Severity: SeverityHigh, Description: "Identify the types of events that the system is capable of logging."},
		{ControlID: "AU-3", Title: "Content of Audit Records", Category: "Audit and Accountability", Severity: SeverityMedium, Description: "Audit records contain information that establishes what type of event occurred."},
		{ControlID: "AU-4", Title: "Audit Storage Capacity", Category: "Audit and Accountability", Severity: SeverityMedium, Description: "Allocate audit record storage capacity in accordance with organizational requirements."},
		{ControlID: "AU-5", Title: "Response to Audit Processing Failures", Category: "Audit and Accountability", Severity: SeverityHigh, Description: "Alert designated personnel in the event of an audit processing failure."},
		{ControlID: "AU-6", Title: "Audit Review, Analysis, and Reporting", Category: "Audit and Accountability", Severity: SeverityMedium, Description: "Review and analyze system audit records for indications of inappropriate activity."},
		{ControlID: "AU-9", Title: "Protection of Audit Information", Category: "Audit and Accountability", Severity: SeverityHigh, Description: "Protect audit information and audit logging tools from unauthorized access."},
		{ControlID: "AU-11", Title: "Audit Record Retention", Category: "Audit and Accountability", Severity: SeverityMedium, Description: "Retain audit records for an organization-defined time period."},
		{ControlID: "AU-12", Title: "Audit Generation", Category: "Audit and Accountability", Severity: SeverityHigh, Description: "Provide audit record generation capability for the events identified."},

		// Configuration Management
		{ControlID: "CM-1", Title: "Policy and Procedures", Category: "Configuration Management", Severity: SeverityLow, Description: "Develop, document, and disseminate configuration management policy."},
		{ControlID: "CM-2", Title: "Baseline Configuration", Category: "Configuration Management", Severity: SeverityHigh, Description: "Develop, document, and maintain a current baseline configuration of the system."},
		{ControlID: "CM-3", Title: "Configuration Change Control", Category: "Configuration Management", Severity: SeverityHigh, Description: "Determine the types of changes to the system that are configuration-controlled."},
		{ControlID: "CM-6", Title: "Configuration Settings", Category: "Configuration Management", Severity: SeverityHigh, Description: "Establish and document configuration settings for components employed within the system."},
		{ControlID: "CM-7", Title: "Least Functionality", Category: "Configuration Management", Severity: SeverityMedium, Description: "Configure the system to provide only essential capabilities."},
		{ControlID: "CM-8", Title: "System Component Inventory", Category: "Configuration Management", Severity: SeverityMedium, Description: "Develop and document an inventory of system components."},

		// System and Communications Protection
		{ControlID: "SC-1", Title: "Policy and Procedures", Category: "System and Communications Protection", Severity: SeverityLow, Description: "Develop, document, and disseminate system and communications protection policy."},
		{ControlID: "SC-7", Title: "Boundary Protection", Category: "System and Communications Protection", Severity: SeverityCritical, Description: "Monitor and control communications at the external managed interfaces."},
		{ControlID: "SC-8", Title: "Transmission Confidentiality and Integrity", Category: "System and Communications Protection", Severity: SeverityHigh, Description: "Protect the confidentiality and integrity of transmitted information."},
		{ControlID: "SC-12", Title: "Cryptographic Key Establishment and Management", Category: "System and Communications Protection", Severity: SeverityHigh, Description: "Establish and manage cryptographic keys when cryptography is employed."},
		{ControlID: "SC-13", Title: "Cryptographic Protection", Category: "System and Communications Protection", Severity: SeverityHigh, Description: "Determine the cryptographic uses and implement cryptographic protection."},
		{ControlID: "SC-28", Title: "Protection of Information at Rest", Category: "System and Communications Protection", Severity: SeverityHigh, Description: "Protect the confidentiality and integrity of information at rest."},

		// System and Information Integrity
		{ControlID: "SI-1", Title: "Policy and Procedures", Category: "System and Information Integrity", Severity: SeverityLow, Description: "Develop, document, and disseminate system and information integrity policy."},
		{ControlID: "SI-2", Title: "Flaw Remediation", Category: "System and Information Integrity", Severity: SeverityHigh, Description: "Identify, report, and correct system flaws in a timely manner."},
		{ControlID: "SI-3", Title: "Malicious Code Protection", Category: "System and Information Integrity", Severity: SeverityHigh, Description: "Implement malicious code protection mechanisms."},
		{ControlID: "SI-4", Title: "System Monitoring", Category: "System and Information Integrity", Severity: SeverityHigh, Description: "Monitor the system to detect attacks and indicators of potential attacks."},
		{ControlID: "SI-5", Title: "Security Alerts, Advisories, and Directives", Category: "System and Information Integrity", Severity: SeverityMedium, Description: "Receive system security alerts, advisories, and directives."},
	}
}

// SOC2Controls returns SOC2 Trust Service Criteria controls
func SOC2Controls() []*Control {
	return []*Control{
		// Common Criteria (CC)
		{ControlID: "CC1.1", Title: "Control Environment", Category: "Common Criteria", Severity: SeverityMedium, Description: "COSO Principle 1: The entity demonstrates a commitment to integrity and ethical values."},
		{ControlID: "CC1.2", Title: "Board Independence", Category: "Common Criteria", Severity: SeverityMedium, Description: "COSO Principle 2: The board of directors demonstrates independence from management."},
		{ControlID: "CC2.1", Title: "Internal Communication", Category: "Common Criteria", Severity: SeverityMedium, Description: "COSO Principle 14: The entity internally communicates information."},
		{ControlID: "CC2.2", Title: "External Communication", Category: "Common Criteria", Severity: SeverityMedium, Description: "COSO Principle 15: The entity communicates with external parties."},
		{ControlID: "CC3.1", Title: "Objective Specification", Category: "Common Criteria", Severity: SeverityMedium, Description: "COSO Principle 6: The entity specifies objectives with sufficient clarity."},
		{ControlID: "CC3.2", Title: "Risk Identification", Category: "Common Criteria", Severity: SeverityHigh, Description: "COSO Principle 7: The entity identifies risks to the achievement of its objectives."},
		{ControlID: "CC4.1", Title: "Change Management", Category: "Common Criteria", Severity: SeverityHigh, Description: "COSO Principle 9: The entity identifies and assesses changes."},
		{ControlID: "CC5.1", Title: "Control Activities", Category: "Common Criteria", Severity: SeverityHigh, Description: "COSO Principle 10: The entity selects and develops control activities."},
		{ControlID: "CC5.2", Title: "Technology Controls", Category: "Common Criteria", Severity: SeverityHigh, Description: "COSO Principle 11: The entity selects and develops general controls over technology."},

		// Logical and Physical Access Controls (CC6)
		{ControlID: "CC6.1", Title: "Logical Access Security", Category: "Logical Access", Severity: SeverityCritical, Description: "The entity implements logical access security software, infrastructure, and architectures."},
		{ControlID: "CC6.2", Title: "Access Registration and Authorization", Category: "Logical Access", Severity: SeverityHigh, Description: "Prior to issuing system credentials, the entity registers and authorizes new users."},
		{ControlID: "CC6.3", Title: "Access Removal", Category: "Logical Access", Severity: SeverityHigh, Description: "The entity removes access to protected information assets when appropriate."},
		{ControlID: "CC6.6", Title: "Encryption", Category: "Logical Access", Severity: SeverityCritical, Description: "The entity implements logical access security measures to protect against threats from sources outside its system boundaries."},
		{ControlID: "CC6.7", Title: "Transmission Protection", Category: "Logical Access", Severity: SeverityHigh, Description: "The entity restricts the transmission, movement, and removal of information."},
		{ControlID: "CC6.8", Title: "Malicious Software Prevention", Category: "Logical Access", Severity: SeverityHigh, Description: "The entity implements controls to prevent or detect and act upon the introduction of malicious software."},

		// System Operations (CC7)
		{ControlID: "CC7.1", Title: "Security Monitoring", Category: "System Operations", Severity: SeverityHigh, Description: "The entity uses detection and monitoring procedures to identify security events."},
		{ControlID: "CC7.2", Title: "Security Event Analysis", Category: "System Operations", Severity: SeverityHigh, Description: "The entity evaluates security events to determine whether they could or have resulted in an incident."},
		{ControlID: "CC7.3", Title: "Incident Response", Category: "System Operations", Severity: SeverityHigh, Description: "The entity responds to identified security incidents."},
		{ControlID: "CC7.4", Title: "Incident Recovery", Category: "System Operations", Severity: SeverityMedium, Description: "The entity responds to identified security incidents by executing recovery procedures."},

		// Change Management (CC8)
		{ControlID: "CC8.1", Title: "Change Authorization", Category: "Change Management", Severity: SeverityHigh, Description: "The entity authorizes, designs, develops, and implements changes to infrastructure, data, software, and procedures."},

		// Risk Mitigation (CC9)
		{ControlID: "CC9.1", Title: "Risk Mitigation", Category: "Risk Mitigation", Severity: SeverityMedium, Description: "The entity identifies, selects, and develops risk mitigation activities."},
		{ControlID: "CC9.2", Title: "Business Risk Management", Category: "Risk Mitigation", Severity: SeverityMedium, Description: "The entity assesses and manages risks associated with vendors and business partners."},

		// Availability
		{ControlID: "A1.1", Title: "Availability Capacity Planning", Category: "Availability", Severity: SeverityMedium, Description: "The entity maintains, monitors, and evaluates current processing capacity."},
		{ControlID: "A1.2", Title: "Environmental Protection", Category: "Availability", Severity: SeverityMedium, Description: "The entity authorizes, designs, develops, and implements environmental protections."},
		{ControlID: "A1.3", Title: "Backup and Recovery", Category: "Availability", Severity: SeverityHigh, Description: "The entity designs, develops, and implements backup and recovery procedures."},

		// Confidentiality
		{ControlID: "C1.1", Title: "Confidential Information Identification", Category: "Confidentiality", Severity: SeverityHigh, Description: "The entity identifies and maintains confidential information."},
		{ControlID: "C1.2", Title: "Confidential Information Disposal", Category: "Confidentiality", Severity: SeverityMedium, Description: "The entity disposes of confidential information."},
	}
}
