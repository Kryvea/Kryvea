import { useEffect, useMemo, useState } from "react";
import Grid from "../../Composition/Grid";
import Accordion from "../../Form/Accordion";
import Input from "../../Form/Input";
import ScoreBar from "../ScoreBar";
import { calculateCVSSFromMetrics, calculateCVSSFromVector } from "./CVSS31";
import CVSS31Render from "./CVSS31Render";

export default function CVSS31Wrapper({ value, onChange }) {
  const [isAccordionOpen, setIsAccordionOpen] = useState(false);
  const [selectedValues, setSelectedValues] = useState({
    AttackVector: "N",
    AttackComplexity: "L",
    PrivilegesRequired: "N",
    UserInteraction: "N",
    Scope: "U",
    Confidentiality: "N",
    Integrity: "N",
    Availability: "N",
    ExploitCodeMaturity: "X",
    RemediationLevel: "X",
    ReportConfidence: "X",
    ConfidentialityRequirement: "X",
    IntegrityRequirement: "X",
    AvailabilityRequirement: "X",
    ModifiedAttackVector: "X",
    ModifiedAttackComplexity: "X",
    ModifiedPrivilegesRequired: "X",
    ModifiedUserInteraction: "X",
    ModifiedScope: "X",
    ModifiedConfidentiality: "X",
    ModifiedIntegrity: "X",
    ModifiedAvailability: "X",
  });
  const [cvssString, setCvssString] = useState(value);
  const [error, setError] = useState("");

  const cvssScore = useMemo(() => {
    const cvssInfo = calculateCVSSFromVector(cvssString);
    if (cvssInfo.success === false) {
      setError(`${cvssInfo.errorType}: ${cvssInfo.errorMetrics ?? "no metrics provided"}`);
      return 0;
    }

    setError("");
    return cvssInfo.environmentalMetricScore;
  }, [cvssString]);

  useEffect(() => {
    if (!value) {
      return;
    }

    const parsedCvss = value.startsWith("CVSS:3.1/") ? value : "CVSS:3.1/" + value;
    const cvssInfo = calculateCVSSFromVector(parsedCvss);

    if (cvssInfo.success === false) {
      setError(`${cvssInfo.errorType}: ${cvssInfo.errorMetrics ?? "no metrics provided"}`);
      return;
    }

    setCvssString(cvssInfo.vectorString);
    setSelectedValues(cvssInfo.metrics);
    setError("");
  }, [value]);

  const handleInputChange = e => {
    const cvssStringChange = e.target.value;
    setCvssString(cvssStringChange);

    const parsedCvss = cvssStringChange.startsWith("CVSS:3.1/") ? cvssStringChange : "CVSS:3.1/" + cvssStringChange;
    const cvssInfo = calculateCVSSFromVector(parsedCvss);
    if (cvssInfo.success === false) {
      setError(`${cvssInfo.errorType}: ${cvssInfo.errorMetrics}`);
      return;
    }
    setCvssString(cvssInfo.vectorString);
    onChange?.(cvssInfo.vectorString);
    setSelectedValues(cvssInfo.metrics);
    setError("");
  };

  const handleButtonClick = (key: string, value: string) => {
    setSelectedValues(prev => {
      const updatedValues = { ...prev, [key]: value };
      const updatedVectorString = calculateCVSSFromMetrics(updatedValues);

      if (updatedVectorString.success === false) {
        setError(`${updatedVectorString.errorType}: ${updatedVectorString.errorMetrics}`);
        return updatedValues;
      }

      setCvssString(updatedVectorString.vectorString);
      onChange?.(updatedVectorString.vectorString);
      return updatedValues;
    });
  };

  return (
    <div className="relative pb-2">
      <Grid
        className={`top-0 grid-cols-[63%_36%] bg-[color:--bg-tertiary] pb-4 pt-2 ${isAccordionOpen ? "sticky z-10" : ""}`}
      >
        <Input
          className={error ? "border-[1px] border-[color:--error]" : ""}
          type="text"
          label="CVSSv3.1 vector"
          id="cvssv3"
          value={cvssString}
          onChange={handleInputChange}
        />
        <ScoreBar score={cvssScore} />
      </Grid>
      <Accordion title={"CVSSv3.1 Calculator"} getIsOpen={setIsAccordionOpen}>
        <CVSS31Render
          {...{
            selectedValues,
            handleButtonClick,
          }}
        />
      </Accordion>
    </div>
  );
}
