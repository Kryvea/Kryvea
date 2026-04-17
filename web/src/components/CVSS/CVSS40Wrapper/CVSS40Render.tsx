import { v4 } from "uuid";
import Grid from "../../Composition/Grid";
import Button from "../../Form/Button";
import Buttons from "../../Form/Buttons";

interface MetricValue {
  name: string;
  description: string;
}

interface MetricConfig {
  label: string;
  group: string;
  subgroup: string | null;
  values: Record<string, MetricValue>;
}

const CVSS_METRICS: Record<string, MetricConfig> = {
  // BASE METRICS - Exploitability
  AttackVector: {
    label: "Attack Vector (AV)",
    group: "base",
    subgroup: "exploitability",
    values: {
      N: {
        name: "Network",
        description:
          'The vulnerable system is bound to the network stack and the set of possible attackers extends beyond the other options listed below, up to and including the entire Internet. Such a vulnerability is often termed "remotely exploitable" and can be thought of as an attack being exploitable at the protocol level one or more network hops away (e.g., across one or more routers).',
      },
      A: {
        name: "Adjacent",
        description:
          "The vulnerable system is bound to a protocol stack, but the attack is limited at the protocol level to a logically adjacent topology. This can mean an attack must be launched from the same shared proximity (e.g., Bluetooth, NFC, or IEEE 802.11) or logical network (e.g., local IP subnet), or from within a secure or otherwise limited administrative domain (e.g., MPLS, secure VPN within an administrative network zone).",
      },
      L: {
        name: "Local",
        description:
          "The vulnerable system is not bound to the network stack and the attacker's path is via read/write/execute capabilities. Either: the attacker exploits the vulnerability by accessing the target system locally (e.g., keyboard, console), or through terminal emulation (e.g., SSH); or the attacker relies on User Interaction by another person to perform actions required to exploit the vulnerability (e.g., using social engineering techniques to trick a legitimate user into opening a malicious document).",
      },
      P: {
        name: "Physical",
        description:
          "The attack requires the attacker to physically touch or manipulate the vulnerable system. Physical interaction may be brief (e.g., evil maid attack) or persistent.",
      },
    },
  },
  AttackComplexity: {
    label: "Attack Complexity (AC)",
    group: "base",
    subgroup: "exploitability",
    values: {
      L: {
        name: "Low",
        description:
          "The attacker must take no measurable action to exploit the vulnerability. The attack requires no target-specific circumvention to exploit the vulnerability. An attacker can expect repeatable success against the vulnerable system.",
      },
      H: {
        name: "High",
        description:
          "The successful attack depends on the evasion or circumvention of security-enhancing techniques in place that would otherwise hinder the attack. These include: Evasion of exploit mitigation techniques. The attacker must have additional methods available to bypass security measures in place. For example, circumvention of address space randomization (ASLR) or data execution prevention (DEP) must be performed for the attack to be successful. Obtaining target-specific secrets. The attacker must gather some target-specific secret before the attack can be successful. A secret is any piece of information that cannot be obtained through any amount of reconnaissance. To obtain the secret the attacker must perform additional attacks or break otherwise secure measures (e.g. knowledge of a secret key may be needed to break a crypto channel). This operation must be performed for each attacked target.",
      },
    },
  },
  AttackRequirements: {
    label: "Attack Requirements (AT)",
    group: "base",
    subgroup: "exploitability",
    values: {
      N: {
        name: "None",
        description:
          "The successful attack does not depend on the deployment and execution conditions of the vulnerable system. The attacker can expect to be able to reach the vulnerability and execute the exploit under all or most instances of the vulnerability.",
      },
      P: {
        name: "Present",
        description:
          "The successful attack depends on the presence of specific deployment and execution conditions of the vulnerable system that enable the attack. These include: A race condition must be won to successfully exploit the vulnerability. The successfulness of the attack is conditioned on execution conditions that are not under full control of the attacker. The attack may need to be launched multiple times against a single target before being successful. Network injection. The attacker must inject themselves into the logical network path between the target and the resource requested by the victim (e.g. vulnerabilities requiring an on-path attacker).",
      },
    },
  },
  PrivilegesRequired: {
    label: "Privileges Required (PR)",
    group: "base",
    subgroup: "exploitability",
    values: {
      N: {
        name: "None",
        description:
          "The attacker is unauthenticated prior to attack, and therefore does not require any access to settings or files of the vulnerable system to carry out an attack.",
      },
      L: {
        name: "Low",
        description:
          "The attacker requires privileges that provide basic capabilities that are typically limited to settings and resources owned by a single low-privileged user. Alternatively, an attacker with Low privileges has the ability to access only non-sensitive resources.",
      },
      H: {
        name: "High",
        description:
          "The attacker requires privileges that provide significant (e.g., administrative) control over the vulnerable system allowing full access to the vulnerable system's settings and files.",
      },
    },
  },
  UserInteraction: {
    label: "User Interaction (UI)",
    group: "base",
    subgroup: "exploitability",
    values: {
      N: {
        name: "None",
        description:
          "The vulnerable system can be exploited without interaction from any human user, other than the attacker. Examples include: a remote attacker is able to send packets to a target system, a locally authenticated attacker executes code to elevate privileges.",
      },
      P: {
        name: "Passive",
        description:
          "Successful exploitation of this vulnerability requires limited interaction by the targeted user with the vulnerable system and the attacker's payload. These interactions would be considered involuntary and do not require that the user actively subvert protections built into the vulnerable system. Examples include: utilizing a website that has been modified to display malicious content when the page is rendered (most stored XSS or CSRF), running an application that calls a malicious binary that has been planted on the system, using an application which generates traffic over an untrusted or compromised network (vulnerabilities requiring an on-path attacker).",
      },
      A: {
        name: "Active",
        description:
          "Successful exploitation of this vulnerability requires a targeted user to perform specific, conscious interactions with the vulnerable system and the attacker's payload, or the user's interactions would actively subvert protection mechanisms which would lead to exploitation of the vulnerability. Examples include: importing a file into a vulnerable system in a specific manner, placing files into a specific directory prior to executing code, submitting a specific string into a web application (e.g. reflected or self XSS), dismiss or accept prompts or security warnings prior to taking an action (e.g. opening/editing a file, connecting a device).",
      },
    },
  },

  // BASE METRICS - Vulnerable System Impact
  Confidentiality: {
    label: "Confidentiality (VC)",
    group: "base",
    subgroup: "vulnerable-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of confidentiality, resulting in all information within the Vulnerable System being divulged to the attacker. Alternatively, access to only some restricted information is obtained, but the disclosed information presents a direct, serious impact. For example, an attacker steals the administrator's password, or private encryption keys of a web server.",
      },
      L: {
        name: "Low",
        description:
          "There is some loss of confidentiality. Access to some restricted information is obtained, but the attacker does not have control over what information is obtained, or the amount or kind of loss is limited. The information disclosure does not cause a direct, serious loss to the Vulnerable System.",
      },
      N: {
        name: "None",
        description: "There is no loss of confidentiality within the Vulnerable System.",
      },
    },
  },
  Integrity: {
    label: "Integrity (VI)",
    group: "base",
    subgroup: "vulnerable-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of integrity, or a complete loss of protection. For example, the attacker is able to modify any/all files protected by the Vulnerable System. Alternatively, only some files can be modified, but malicious modification would present a direct, serious consequence to the Vulnerable System.",
      },
      L: {
        name: "Low",
        description:
          "Modification of data is possible, but the attacker does not have control over the consequence of a modification, or the amount of modification is limited. The data modification does not have a direct, serious impact to the Vulnerable System.",
      },
      N: {
        name: "None",
        description: "There is no loss of integrity within the Vulnerable System.",
      },
    },
  },
  Availability: {
    label: "Availability (VA)",
    group: "base",
    subgroup: "vulnerable-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of availability, resulting in the attacker being able to fully deny access to resources in the Vulnerable System; this loss is either sustained (while the attacker continues to deliver the attack) or persistent (the condition persists even after the attack has completed). Alternatively, the attacker has the ability to deny some availability, but the loss of availability presents a direct, serious consequence to the Vulnerable System (e.g., the attacker cannot disrupt existing connections, but can prevent new connections; the attacker can repeatedly exploit a vulnerability that, in each instance of a successful attack, leaks a only small amount of memory, but after repeated exploitation causes a service to become completely unavailable).",
      },
      L: {
        name: "Low",
        description:
          "Performance is reduced or there are interruptions in resource availability. Even if repeated exploitation of the vulnerability is possible, the attacker does not have the ability to completely deny service to legitimate users. The resources in the Vulnerable System are either partially available all of the time, or fully available only some of the time, but overall there is no direct, serious consequence to the Vulnerable System.",
      },
      N: {
        name: "None",
        description: "There is no impact to availability within the Vulnerable System.",
      },
    },
  },

  // BASE METRICS - Subsequent System Impact
  SubsequentConfidentiality: {
    label: "Confidentiality (SC)",
    group: "base",
    subgroup: "subsequent-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of confidentiality, resulting in all resources within the Subsequent System being divulged to the attacker. Alternatively, access to only some restricted information is obtained, but the disclosed information presents a direct, serious impact. For example, an attacker steals the administrator's password, or private encryption keys of a web server.",
      },
      L: {
        name: "Low",
        description:
          "There is some loss of confidentiality. Access to some restricted information is obtained, but the attacker does not have control over what information is obtained, or the amount or kind of loss is limited. The information disclosure does not cause a direct, serious loss to the Subsequent System.",
      },
      N: {
        name: "None",
        description:
          "There is no loss of confidentiality within the Subsequent System or all confidentiality impact is constrained to the Vulnerable System.",
      },
    },
  },
  SubsequentIntegrity: {
    label: "Integrity (SI)",
    group: "base",
    subgroup: "subsequent-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of integrity, or a complete loss of protection. For example, the attacker is able to modify any/all files protected by the Subsequent System. Alternatively, only some files can be modified, but malicious modification would present a direct, serious consequence to the Subsequent System.",
      },
      L: {
        name: "Low",
        description:
          "Modification of data is possible, but the attacker does not have control over the consequence of a modification, or the amount of modification is limited. The data modification does not have a direct, serious impact to the Subsequent System.",
      },
      N: {
        name: "None",
        description:
          "There is no loss of integrity within the Subsequent System or all integrity impact is constrained to the Vulnerable System.",
      },
    },
  },
  SubsequentAvailability: {
    label: "Availability (SA)",
    group: "base",
    subgroup: "subsequent-impact",
    values: {
      H: {
        name: "High",
        description:
          "There is a total loss of availability, resulting in the attacker being able to fully deny access to resources in the Subsequent System; this loss is either sustained (while the attacker continues to deliver the attack) or persistent (the condition persists even after the attack has completed). Alternatively, the attacker has the ability to deny some availability, but the loss of availability presents a direct, serious consequence to the Subsequent System (e.g., the attacker cannot disrupt existing connections, but can prevent new connections; the attacker can repeatedly exploit a vulnerability that, in each instance of a successful attack, leaks a only small amount of memory, but after repeated exploitation causes a service to become completely unavailable).",
      },
      L: {
        name: "Low",
        description:
          "Performance is reduced or there are interruptions in resource availability. Even if repeated exploitation of the vulnerability is possible, the attacker does not have the ability to completely deny service to legitimate users. The resources in the Subsequent System are either partially available all of the time, or fully available only some of the time, but overall there is no direct, serious consequence to the Subsequent System.",
      },
      N: {
        name: "None",
        description:
          "There is no impact to availability within the Subsequent System or all availability impact is constrained to the Vulnerable System.",
      },
    },
  },

  // THREAT METRICS
  ExploitMaturity: {
    label: "Exploit Maturity (E)",
    group: "threat",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description:
          "The Exploit Maturity metric is not being used. Reliable threat intelligence is not available to determine Exploit Maturity characteristics.",
      },
      A: {
        name: "Attacked",
        description:
          "Based on available threat intelligence either of the following must apply: Attacks targeting this vulnerability (attempted or successful) have been reported. Solutions to simplify attempts to exploit the vulnerability are publicly or privately available (such as exploit toolkits).",
      },
      P: {
        name: "Proof-of-Concept",
        description:
          'Based on available threat intelligence each of the following must apply: Proof-of-concept exploit code is publicly available. No knowledge of reported attempts to exploit this vulnerability. No knowledge of publicly available solutions used to simplify attempts to exploit the vulnerability (i.e., the "Attacked" value does not apply).',
      },
      U: {
        name: "Unreported",
        description:
          'Based on available threat intelligence each of the following must apply: No knowledge of publicly available proof-of-concept exploit code. No knowledge of reported attempts to exploit this vulnerability. No knowledge of publicly available solutions used to simplify attempts to exploit the vulnerability (i.e., neither the "POC" nor "Attacked" values apply).',
      },
    },
  },

  // ENVIRONMENTAL METRICS - Security Requirements
  ConfidentialityRequirements: {
    label: "Confidentiality Requirements (CR)",
    group: "environmental",
    subgroup: "security-requirements",
    values: {
      X: {
        name: "Not Defined",
        description:
          "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score.",
      },
      H: {
        name: "High",
        description:
          "Loss of Confidentiality is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
      },
      M: {
        name: "Medium",
        description:
          "Loss of Confidentiality is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
      },
      L: {
        name: "Low",
        description:
          "Loss of Confidentiality is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
      },
    },
  },
  IntegrityRequirements: {
    label: "Integrity Requirements (IR)",
    group: "environmental",
    subgroup: "security-requirements",
    values: {
      X: {
        name: "Not Defined",
        description:
          "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score.",
      },
      H: {
        name: "High",
        description:
          "Loss of Integrity is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
      },
      M: {
        name: "Medium",
        description:
          "Loss of Integrity is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
      },
      L: {
        name: "Low",
        description:
          "Loss of Integrity is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
      },
    },
  },
  AvailabilityRequirements: {
    label: "Availability Requirements (AR)",
    group: "environmental",
    subgroup: "security-requirements",
    values: {
      X: {
        name: "Not Defined",
        description:
          "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score",
      },
      H: {
        name: "High",
        description:
          "Loss of Availability is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
      },
      M: {
        name: "Medium",
        description:
          "Loss of Availability is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
      },
      L: {
        name: "Low",
        description:
          "Loss of Availability is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
      },
    },
  },

  // ENVIRONMENTAL METRICS - Modified Base (Exploitability)
  ModifiedAttackVector: {
    label: "Attack Vector (MAV)",
    group: "environmental",
    subgroup: "modified-exploitability",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "Network",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      A: {
        name: "Adjacent",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Local",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      P: {
        name: "Physical",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedAttackComplexity: {
    label: "Attack Complexity (MAC)",
    group: "environmental",
    subgroup: "modified-exploitability",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedAttackRequirements: {
    label: "Attack Requirements (MAT)",
    group: "environmental",
    subgroup: "modified-exploitability",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      P: {
        name: "Present",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedPrivilegesRequired: {
    label: "Privileges Required (MPR)",
    group: "environmental",
    subgroup: "modified-exploitability",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedUserInteraction: {
    label: "User Interaction (MUI)",
    group: "environmental",
    subgroup: "modified-exploitability",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      P: {
        name: "Passive",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      A: {
        name: "Active",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },

  // ENVIRONMENTAL METRICS - Modified Vulnerable System Impact
  ModifiedConfidentiality: {
    label: "Confidentiality (MVC)",
    group: "environmental",
    subgroup: "modified-vulnerable-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedIntegrity: {
    label: "Integrity (MVI)",
    group: "environmental",
    subgroup: "modified-vulnerable-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedAvailability: {
    label: "Availability (MVA)",
    group: "environmental",
    subgroup: "modified-vulnerable-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "None",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },

  // ENVIRONMENTAL METRICS - Modified Subsequent System Impact
  ModifiedSubsequentConfidentiality: {
    label: "Confidentiality (MSC)",
    group: "environmental",
    subgroup: "modified-subsequent-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "Negligible",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedSubsequentIntegrity: {
    label: "Integrity (MSI)",
    group: "environmental",
    subgroup: "modified-subsequent-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      S: {
        name: "Safety",
        description:
          "The system may have safety implications as a matter of how or where it is deployed. If the exploitation of a technical vulnerability has the potential to impact human safety, the modified subsequent system impact of Safety should be used.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "Negligible",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },
  ModifiedSubsequentAvailability: {
    label: "Availability (MSA)",
    group: "environmental",
    subgroup: "modified-subsequent-impact",
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      S: {
        name: "Safety",
        description:
          "The system may have safety implications as a matter of how or where it is deployed. If the exploitation of a technical vulnerability has the potential to impact human safety, the modified subsequent system impact of Safety should be used.",
      },
      H: {
        name: "High",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      L: {
        name: "Low",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
      N: {
        name: "Negligible",
        description: "This metric value has the same definition as the Base Metric value defined above.",
      },
    },
  },

  // SUPPLEMENTAL METRICS
  Safety: {
    label: "Safety (S)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "Negligible",
        description:
          'Consequences of the vulnerability meet definition of IEC 61508 consequence category "negligible."',
      },
      P: {
        name: "Present",
        description:
          'Consequences of the vulnerability meet definition of IEC 61508 consequence categories of "marginal," "critical," or "catastrophic."',
      },
    },
  },
  Automatable: {
    label: "Automatable (AU)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      N: {
        name: "No",
        description:
          "Attackers cannot reliably automate all 4 steps of the kill chain for this vulnerability for some reason. These steps are reconnaissance, weaponization, delivery, and exploitation.",
      },
      Y: {
        name: "Yes",
        description:
          'Attackers can reliably automate all 4 steps of the kill chain. These steps are reconnaissance, weaponization, delivery, and exploitation (e.g., the vulnerability is "wormable").',
      },
    },
  },
  Recovery: {
    label: "Recovery (R)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      A: {
        name: "Automatic",
        description: "The system recovers services automatically after an attack has been performed.",
      },
      U: {
        name: "User",
        description:
          "The system requires manual intervention by the user to recover services, after an attack has been performed.",
      },
      I: {
        name: "Irrecoverable",
        description: "The system services are irrecoverable by the user, after an attack has been performed.",
      },
    },
  },
  ValueDensity: {
    label: "Value Density (V)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      D: {
        name: "Diffuse",
        description:
          "The vulnerable system has limited resources. That is, the resources that the attacker will gain control over with a single exploitation event are relatively small. An example of Diffuse (think: limited) Value Density would be an attack on a single email client vulnerability.",
      },
      C: {
        name: "Concentrated",
        description:
          'The vulnerable system is rich in resources. Heuristically, such systems are often the direct responsibility of "system operators" rather than users. An example of Concentrated (think: broad) Value Density would be an attack on a central email server.',
      },
    },
  },
  ResponseEffort: {
    label: "Vulnerability Response Effort (RE)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      L: {
        name: "Low",
        description:
          "The effort required to respond to a vulnerability is low/trivial. Examples include: communication on better documentation, configuration workarounds, or guidance from the vendor that does not require an immediate update, upgrade, or replacement by the consuming entity, such as firewall filter configuration.",
      },
      M: {
        name: "Moderate",
        description:
          "The actions required to respond to a vulnerability require some effort on behalf of the consumer and could cause minimal service impact to implement. Examples include: simple remote update, disabling of a subsystem, or a low-touch software upgrade such as a driver update.",
      },
      H: {
        name: "High",
        description:
          "The actions required to respond to a vulnerability are significant and/or difficult, and may possibly lead to an extended, scheduled service impact. This would need to be considered for scheduling purposes including honoring any embargo on deployment of the selected response. Alternatively, response to the vulnerability in the field is not possible remotely. The only resolution to the vulnerability involves physical replacement (e.g. units deployed would have to be recalled for a depot level repair or replacement). Examples include: a highly privileged driver update, microcode or UEFI BIOS updates, or software upgrades requiring careful analysis and understanding of any potential infrastructure impact before implementation.",
      },
    },
  },
  ProviderUrgency: {
    label: "Provider Urgency (U)",
    group: "supplemental",
    subgroup: null,
    values: {
      X: {
        name: "Not Defined",
        description: "The metric has not been evaluated.",
      },
      Red: {
        name: "Red",
        description: "Provider has assessed the impact of this vulnerability as having the highest urgency.",
      },
      Amber: {
        name: "Amber",
        description: "Provider has assessed the impact of this vulnerability as having a moderate urgency.",
      },
      Green: {
        name: "Green",
        description: "Provider has assessed the impact of this vulnerability as having a reduced urgency.",
      },
      Clear: {
        name: "Clear",
        description: "Provider has assessed the impact of this vulnerability as having no urgency (Informational).",
      },
    },
  },
};

export default function CVSS40Render({ selectedValues, handleButtonClick }) {
  // Helper function to get metrics by group and subgroup
  const getMetricsByFilter = (group: string, subgroup?: string | null) => {
    return Object.entries(CVSS_METRICS).filter(([_, config]) => {
      if (subgroup !== undefined) {
        return config.group === group && config.subgroup === subgroup;
      }
      return config.group === group;
    });
  };

  const renderMetricButtons = ([metricKey, metricConfig]: [string, MetricConfig]) => {
    const cvss40KeyUuid = v4();
    return (
      <div className="pt-2" key={`container-${cvss40KeyUuid}`}>
        <Buttons label={metricConfig.label}>
          {Object.entries(metricConfig.values).map(([optionKey, optionData]: [string, MetricValue]) => {
            const isSelected = optionKey === selectedValues[metricKey];
            return (
              <Button
                small
                variant={isSelected ? "" : "secondary"}
                text={`${optionData.name} (${optionKey})`}
                onClick={() => handleButtonClick(metricKey, optionKey)}
                key={`${optionKey}-${cvss40KeyUuid}`}
                title={optionData.description}
              />
            );
          })}
        </Buttons>
      </div>
    );
  };

  return (
    <Grid className="pt-4">
      {/* Base Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Base Metrics</h3>
          <div className="grid grid-cols-3">
            {/* Exploitability Metrics */}
            <div>
              <h4 className="text-xl font-bold">Exploitability Metrics</h4>
              {getMetricsByFilter("base", "exploitability").map(renderMetricButtons)}
            </div>
            {/* Vulnerable System Impact Metrics */}
            <div>
              <h4 className="text-xl font-bold">Vulnerable System Impact Metrics</h4>
              {getMetricsByFilter("base", "vulnerable-impact").map(renderMetricButtons)}
            </div>
            {/* Subsequent System Impact Metrics */}
            <div>
              <h4 className="text-xl font-bold">Subsequent System Impact Metrics</h4>
              {getMetricsByFilter("base", "subsequent-impact").map(renderMetricButtons)}
            </div>
          </div>
        </div>
      </Grid>

      {/* Supplemental Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Supplemental Metrics</h3>
          <div className="grid grid-cols-2">
            <div>{getMetricsByFilter("supplemental").slice(0, 3).map(renderMetricButtons)}</div>
            <div>{getMetricsByFilter("supplemental").slice(3).map(renderMetricButtons)}</div>
          </div>
        </div>
      </Grid>

      {/* Environmental (Modified Base Metrics) */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Environmental (Modified Base Metrics)</h3>
          <div className="grid grid-cols-3">
            {/* Exploitability Metrics */}
            <div>
              <h4 className="text-xl font-bold">Exploitability Metrics</h4>
              {getMetricsByFilter("environmental", "modified-exploitability").map(renderMetricButtons)}
            </div>
            {/* Vulnerable System Impact Metrics */}
            <div>
              <h4 className="text-xl font-bold">Vulnerable System Impact Metrics</h4>
              {getMetricsByFilter("environmental", "modified-vulnerable-impact").map(renderMetricButtons)}
            </div>
            {/* Subsequent System Impact Metrics */}
            <div>
              <h4 className="text-xl font-bold">Subsequent System Impact Metrics</h4>
              {getMetricsByFilter("environmental", "modified-subsequent-impact").map(renderMetricButtons)}
            </div>
          </div>
        </div>
      </Grid>

      {/* Environmental (Security Requirements) */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Environmental (Security Requirements)</h3>
          <Grid className="grid-cols-3">
            {getMetricsByFilter("environmental", "security-requirements").map(renderMetricButtons)}
          </Grid>
        </div>
      </Grid>

      {/* Threat Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Threat Metrics</h3>
          <div className="flex flex-col">{getMetricsByFilter("threat").map(renderMetricButtons)}</div>
        </div>
      </Grid>
    </Grid>
  );
}
