import { v4 } from "uuid";
import Grid from "../../Composition/Grid";
import Button from "../../Form/Button";
import Buttons from "../../Form/Buttons";

interface MetricValue {
  name: string;
  description: string;
}

interface Metric {
  label: string;
  values: Record<string, MetricValue>;
}

interface Metrics {
  [key: string]: Metric;
}

interface CVSS31RenderProps {
  selectedValues: Record<string, string>;
  handleButtonClick: (metricKey: string, optionKey: string) => void;
}

export default function CVSS31Render({ selectedValues, handleButtonClick }: CVSS31RenderProps) {
  const metrics: Metrics = {
    AttackVector: {
      label: "Attack Vector (AV)",
      values: {
        N: {
          name: "Network",
          description:
            "The vulnerable component is bound to the network stack and the attacker's path is through the network layer. Such a vulnerability is often termed 'remotely exploitable'.",
        },
        A: {
          name: "Adjacent Network",
          description:
            "The vulnerable component is bound to the network stack, but the attack is limited at the protocol level to a logically adjacent topology. This can mean an attack must be launched from the same shared physical or logical network.",
        },
        L: {
          name: "Local",
          description:
            "The vulnerable component is not bound to the network stack and the attacker's path is via read/write/execute capabilities. Either the attacker exploits the vulnerability by accessing the target system locally or remotely, or the attacker relies on User Interaction by another person to perform actions required to exploit the vulnerability.",
        },
        P: {
          name: "Physical",
          description:
            "The attack requires the attacker to physically touch or manipulate the vulnerable component. Physical interaction may be brief or persistent.",
        },
      },
    },
    AttackComplexity: {
      label: "Attack Complexity (AC)",
      values: {
        L: {
          name: "Low",
          description:
            "Specialized access conditions or extenuating circumstances do not exist. An attacker can expect repeatable success when attacking the vulnerable component.",
        },
        H: {
          name: "High",
          description:
            "A successful attack depends on conditions beyond the attacker's control. That is, a successful attack cannot be accomplished at will, but requires considerable preparation or execution against the specific target.",
        },
      },
    },
    PrivilegesRequired: {
      label: "Privileges Required (PR)",
      values: {
        N: {
          name: "None",
          description:
            "The attacker is unauthorized prior to attack, and therefore does not require any access to settings or files of the vulnerable system to carry out an attack.",
        },
        L: {
          name: "Low",
          description:
            "The attacker requires privileges that provide basic user capabilities that could normally affect only settings and files owned by a user. Alternatively, an attacker with Low privileges has the ability to access only non-sensitive resources.",
        },
        H: {
          name: "High",
          description:
            "The attacker requires privileges that provide significant control over the vulnerable component allowing access to component-wide settings and files.",
        },
      },
    },
    UserInteraction: {
      label: "User Interaction (UI)",
      values: {
        N: {
          name: "None",
          description: "The vulnerable system can be exploited without interaction from any user.",
        },
        R: {
          name: "Required",
          description:
            "Successful exploitation of this vulnerability requires a user to take some action before the vulnerability can be exploited.",
        },
      },
    },
    Scope: {
      label: "Scope (S)",
      values: {
        U: {
          name: "Unchanged",
          description:
            "An exploited vulnerability can only affect resources managed by the same security authority. In this case, the vulnerable component and the impacted component are either the same, or both are managed by the same security authority.",
        },
        C: {
          name: "Changed",
          description:
            "An exploited vulnerability can affect resources beyond the security scope managed by the security authority of the vulnerable component. In this case, the vulnerable component and the impacted component are different and managed by different security authorities.",
        },
      },
    },
    Confidentiality: {
      label: "Confidentiality (C)",
      values: {
        N: {
          name: "None",
          description: "There is no loss of confidentiality within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "There is some loss of confidentiality. Access to some restricted information is obtained, but the attacker does not have control over what information is obtained, or the amount or kind of loss is limited.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of confidentiality, resulting in all resources within the impacted component being divulged to the attacker. Alternatively, access to only some restricted information is obtained, but the disclosed information presents a direct, serious impact.",
        },
      },
    },
    Integrity: {
      label: "Integrity (I)",
      values: {
        N: {
          name: "None",
          description: "There is no loss of integrity within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "Modification of data is possible, but the attacker does not have control over the consequence of a modification, or the amount of modification is limited.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of integrity, or a complete loss of protection. The attacker is able to modify any/all files protected by the impacted component. Alternatively, only some files can be modified, but malicious modification would present a direct, serious consequence to the impacted component.",
        },
      },
    },
    Availability: {
      label: "Availability (A)",
      values: {
        N: {
          name: "None",
          description: "There is no impact to availability within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "Performance is reduced or there are interruptions in resource availability. Even if repeated exploitation of the vulnerability is possible, the attacker does not have the ability to completely deny service to legitimate users.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of availability, resulting in the attacker being able to fully deny access to resources in the impacted component; this loss is either sustained or persistent.",
        },
      },
    },
    ExploitCodeMaturity: {
      label: "Exploit Code Maturity (E)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Temporal Score.",
        },
        U: {
          name: "Unproven",
          description: "No exploit code is available, or an exploit is theoretical.",
        },
        P: {
          name: "Proof of Concept",
          description:
            "Proof-of-concept exploit code is available, or an attack demonstration is not practical for most systems. The code or technique is not functional in all situations and may require substantial modification by a skilled attacker.",
        },
        F: {
          name: "Functional",
          description:
            "Functional exploit code is available. The code works in most situations where the vulnerability exists.",
        },
        H: {
          name: "High",
          description:
            "Functional autonomous code exists, or no exploit is required (manual trigger) and details are widely available. Exploit code works in every situation, or is actively being delivered via an autonomous agent.",
        },
      },
    },
    RemediationLevel: {
      label: "Remediation Level (RL)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Temporal Score.",
        },
        O: {
          name: "Official Fix",
          description:
            "A complete vendor solution is available. Either the vendor has issued an official patch, or an upgrade is available.",
        },
        T: {
          name: "Temporary Fix",
          description:
            "There is an official but temporary fix available. This includes instances where the vendor issues a temporary hotfix, tool, or workaround.",
        },
        W: {
          name: "Workaround",
          description:
            "There is an unofficial, non-vendor solution available. In some cases, users of the affected technology will create a patch of their own or provide steps to work around or otherwise mitigate the vulnerability.",
        },
        U: {
          name: "Unavailable",
          description: "There is either no solution available or it is impossible to apply.",
        },
      },
    },
    ReportConfidence: {
      label: "Report Confidence (RC)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Temporal Score.",
        },
        U: {
          name: "Unknown",
          description:
            "There are reports of impacts that indicate a vulnerability is present. The reports indicate that the cause of the vulnerability is unknown, or reports may differ on the cause or impacts of the vulnerability.",
        },
        R: {
          name: "Reasonable",
          description:
            "Significant details are published, but researchers either do not have full confidence in the root cause, or do not have access to source code to fully confirm all of the interactions that may lead to the result.",
        },
        C: {
          name: "Confirmed",
          description:
            "Detailed reports exist, or functional reproduction is possible. Source code is available to independently verify the assertions of the research, or the author or vendor of the affected code has confirmed the presence of the vulnerability.",
        },
      },
    },
    ModifiedAttackVector: {
      label: "Modified Attack Vector (MAV)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "Network",
          description:
            "The vulnerable component is bound to the network stack and the attacker's path is through the network layer. Such a vulnerability is often termed 'remotely exploitable'.",
        },
        A: {
          name: "Adjacent Network",
          description:
            "The vulnerable component is bound to the network stack, but the attack is limited at the protocol level to a logically adjacent topology. This can mean an attack must be launched from the same shared physical or logical network.",
        },
        L: {
          name: "Local",
          description:
            "The vulnerable component is not bound to the network stack and the attacker's path is via read/write/execute capabilities. Either the attacker exploits the vulnerability by accessing the target system locally or remotely, or the attacker relies on User Interaction by another person to perform actions required to exploit the vulnerability.",
        },
        P: {
          name: "Physical",
          description:
            "The attack requires the attacker to physically touch or manipulate the vulnerable component. Physical interaction may be brief or persistent.",
        },
      },
    },
    ModifiedAttackComplexity: {
      label: "Modified Attack Complexity (MAC)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        L: {
          name: "Low",
          description:
            "Specialized access conditions or extenuating circumstances do not exist. An attacker can expect repeatable success when attacking the vulnerable component.",
        },
        H: {
          name: "High",
          description:
            "A successful attack depends on conditions beyond the attacker's control. That is, a successful attack cannot be accomplished at will, but requires considerable preparation or execution against the specific target.",
        },
      },
    },
    ModifiedPrivilegesRequired: {
      label: "Modified Privileges Required (MPR)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "None",
          description:
            "The attacker is unauthorized prior to attack, and therefore does not require any access to settings or files of the vulnerable system to carry out an attack.",
        },
        L: {
          name: "Low",
          description:
            "The attacker requires privileges that provide basic user capabilities that could normally affect only settings and files owned by a user. Alternatively, an attacker with Low privileges has the ability to access only non-sensitive resources.",
        },
        H: {
          name: "High",
          description:
            "The attacker requires privileges that provide significant control over the vulnerable component allowing access to component-wide settings and files.",
        },
      },
    },
    ModifiedUserInteraction: {
      label: "Modified User Interaction (MUI)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "None",
          description: "The vulnerable system can be exploited without interaction from any user.",
        },
        R: {
          name: "Required",
          description:
            "Successful exploitation of this vulnerability requires a user to take some action before the vulnerability can be exploited.",
        },
      },
    },
    ModifiedScope: {
      label: "Modified Scope (MS)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        U: {
          name: "Unchanged",
          description:
            "An exploited vulnerability can only affect resources managed by the same security authority. In this case, the vulnerable component and the impacted component are either the same, or both are managed by the same security authority.",
        },
        C: {
          name: "Changed",
          description:
            "An exploited vulnerability can affect resources beyond the security scope managed by the security authority of the vulnerable component. In this case, the vulnerable component and the impacted component are different and managed by different security authorities.",
        },
      },
    },
    ModifiedConfidentiality: {
      label: "Modified Confidentiality (MC)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "None",
          description: "There is no loss of confidentiality within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "There is some loss of confidentiality. Access to some restricted information is obtained, but the attacker does not have control over what information is obtained, or the amount or kind of loss is limited.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of confidentiality, resulting in all resources within the impacted component being divulged to the attacker. Alternatively, access to only some restricted information is obtained, but the disclosed information presents a direct, serious impact.",
        },
      },
    },
    ModifiedIntegrity: {
      label: "Modified Integrity (MI)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "None",
          description: "There is no loss of integrity within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "Modification of data is possible, but the attacker does not have control over the consequence of a modification, or the amount of modification is limited.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of integrity, or a complete loss of protection. The attacker is able to modify any/all files protected by the impacted component. Alternatively, only some files can be modified, but malicious modification would present a direct, serious consequence to the impacted component.",
        },
      },
    },
    ModifiedAvailability: {
      label: "Modified Availability (MA)",
      values: {
        X: {
          name: "Not Defined",
          description: "Use the value assigned to the corresponding Base Score metric.",
        },
        N: {
          name: "None",
          description: "There is no impact to availability within the impacted component.",
        },
        L: {
          name: "Low",
          description:
            "Performance is reduced or there are interruptions in resource availability. Even if repeated exploitation of the vulnerability is possible, the attacker does not have the ability to completely deny service to legitimate users.",
        },
        H: {
          name: "High",
          description:
            "There is a total loss of availability, resulting in the attacker being able to fully deny access to resources in the impacted component; this loss is either sustained or persistent.",
        },
      },
    },
    ConfidentialityRequirement: {
      label: "Confidentiality Requirement (CR)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score.",
        },
        L: {
          name: "Low",
          description:
            "Loss of confidentiality is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
        },
        M: {
          name: "Medium",
          description:
            "Loss of confidentiality is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
        },
        H: {
          name: "High",
          description:
            "Loss of confidentiality is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
        },
      },
    },
    IntegrityRequirement: {
      label: "Integrity Requirement (IR)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score.",
        },
        L: {
          name: "Low",
          description:
            "Loss of integrity is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
        },
        M: {
          name: "Medium",
          description:
            "Loss of integrity is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
        },
        H: {
          name: "High",
          description:
            "Loss of integrity is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
        },
      },
    },
    AvailabilityRequirement: {
      label: "Availability Requirement (AR)",
      values: {
        X: {
          name: "Not Defined",
          description:
            "Assigning this value indicates there is insufficient information to choose one of the other values, and has no impact on the overall Environmental Score.",
        },
        L: {
          name: "Low",
          description:
            "Loss of availability is likely to have only a limited adverse effect on the organization or individuals associated with the organization.",
        },
        M: {
          name: "Medium",
          description:
            "Loss of availability is likely to have a serious adverse effect on the organization or individuals associated with the organization.",
        },
        H: {
          name: "High",
          description:
            "Loss of availability is likely to have a catastrophic adverse effect on the organization or individuals associated with the organization.",
        },
      },
    },
  };

  const metricKeys = Object.keys(metrics);

  return (
    <Grid className="pt-4">
      {/* Base Score Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Base Score Metrics</h3>
          <Grid className="grid-cols-2">
            <div>{metricKeys.slice(0, 4).map(renderMetricButtons)}</div>
            <div>{metricKeys.slice(4, 8).map(renderMetricButtons)}</div>
          </Grid>
        </div>
      </Grid>

      {/* Temporal Score Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Temporal Score Metrics</h3>
          {metricKeys.slice(8, 11).map(renderMetricButtons)}
        </div>
      </Grid>

      {/* Environmental Score Metrics */}
      <Grid>
        <div className="rounded-2xl border border-[color:--border-primary] p-4">
          <h3 className="text-2xl font-bold">Environmental Score Metrics</h3>
          <div className="grid grid-cols-[1fr_2.5rem_1fr_1fr]">
            {/* Exploitability Metrics */}
            <div>
              <h4 className="pt-2 text-xl font-bold">Exploitability Metrics</h4>
              {metricKeys.slice(11, 16).map(renderMetricButtons)}
            </div>

            {/* Spacer column */}
            <div></div>

            {/* Impact Metrics */}
            <div>
              <h4 className="pt-2 text-xl font-bold">Impact Metrics</h4>
              {metricKeys.slice(16, 19).map(renderMetricButtons)}
            </div>

            {/* Impact Subscore Modifiers */}
            <div>
              <h4 className="pt-2 text-xl font-bold">Impact Subscore Modifiers</h4>
              <div>{metricKeys.slice(19).map(renderMetricButtons)}</div>
            </div>
          </div>
        </div>
      </Grid>
    </Grid>
  );

  function renderMetricButtons(metricKey: string) {
    const cvss31KeyUuid = v4();
    const metric = metrics[metricKey];

    return (
      <div className="pt-2" key={`container-${cvss31KeyUuid}`}>
        <Buttons label={metric.label}>
          {Object.entries(metric.values).map(([optionKey, optionData]) => (
            <Button
              small
              variant={optionKey === selectedValues[metricKey] ? "" : "secondary"}
              text={`${optionData.name} (${optionKey})`}
              onClick={() => handleButtonClick(metricKey, optionKey)}
              key={`${optionData.name}-${cvss31KeyUuid}`}
              title={optionData.description}
            />
          ))}
        </Buttons>
      </div>
    );
  }
}
