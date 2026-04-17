import { useEffect, useMemo, useState } from "react";
import Grid from "../../Composition/Grid";
import Accordion from "../../Form/Accordion";
import Input from "../../Form/Input";
import ScoreBar from "../ScoreBar";
import Vector, { CVSS40 } from "./CVSS40";
import CVSS40Render from "./CVSS40Render";

type NALP = "N" | "A" | "L" | "P";
type LH = "L" | "H";
type NP = "N" | "P";
type NPA = "N" | "P" | "A";
type NLH = "N" | "L" | "H";
type XUPA = "X" | "U" | "P" | "A";
type XLMH = "X" | "L" | "M" | "H";
type XNALP = "X" | "N" | "A" | "L" | "P";
type XLH = "X" | "L" | "H";
type XNP = "X" | "N" | "P";
type XNLH = "X" | "N" | "L" | "H";
type XNPA = "X" | "N" | "P" | "A";
type XNLHS = "X" | "N" | "L" | "H" | "S";
type XNY = "X" | "N" | "Y";
type XAUI = "X" | "A" | "U" | "I";
type XDC = "X" | "D" | "C";
type XClearGreenAmberRed = "X" | "Clear" | "Green" | "Amber" | "Red";
type Metrics = {
  AttackVector: NALP;
  AttackComplexity: LH;
  AttackRequirements: NP;
  PrivilegesRequired: NLH;
  UserInteraction: NPA;
  Confidentiality: NLH;
  Integrity: NLH;
  Availability: NLH;
  SubsequentConfidentiality: NLH;
  SubsequentIntegrity: NLH;
  SubsequentAvailability: NLH;
  ExploitMaturity: XUPA;
  ConfidentialityRequirements: XLMH;
  IntegrityRequirements: XLMH;
  AvailabilityRequirements: XLMH;
  ModifiedAttackVector: XNALP;
  ModifiedAttackComplexity: XLH;
  ModifiedAttackRequirements: XNP;
  ModifiedPrivilegesRequired: XNLH;
  ModifiedUserInteraction: XNPA;
  ModifiedConfidentiality: XNLH;
  ModifiedIntegrity: XNLH;
  ModifiedAvailability: XNLH;
  ModifiedSubsequentConfidentiality: XNLH;
  ModifiedSubsequentIntegrity: XNLHS;
  ModifiedSubsequentAvailability: XNLHS;
  Safety: XNP;
  Automatable: XNY;
  Recovery: XAUI;
  ValueDensity: XDC;
  ResponseEffort: XLMH;
  ProviderUrgency: XClearGreenAmberRed;
};

export default function CVSS40Wrapper({ value, onChange }) {
  const [isAccordionOpen, setIsAccordionOpen] = useState(false);
  const [metrics, setMetrics] = useState({
    AttackVector: "N",
    AttackComplexity: "L",
    AttackRequirements: "N",
    PrivilegesRequired: "N",
    UserInteraction: "N",
    Confidentiality: "N",
    Integrity: "N",
    Availability: "N",
    SubsequentConfidentiality: "N",
    SubsequentIntegrity: "N",
    SubsequentAvailability: "N",
    ExploitMaturity: "X",
    ConfidentialityRequirements: "X",
    IntegrityRequirements: "X",
    AvailabilityRequirements: "X",
    ModifiedAttackVector: "X",
    ModifiedAttackComplexity: "X",
    ModifiedAttackRequirements: "X",
    ModifiedPrivilegesRequired: "X",
    ModifiedUserInteraction: "X",
    ModifiedConfidentiality: "X",
    ModifiedIntegrity: "X",
    ModifiedAvailability: "X",
    ModifiedSubsequentConfidentiality: "X",
    ModifiedSubsequentIntegrity: "X",
    ModifiedSubsequentAvailability: "X",
    Safety: "X",
    Automatable: "X",
    Recovery: "X",
    ValueDensity: "X",
    ResponseEffort: "X",
    ProviderUrgency: "X",
  });
  const metricLabelsShort = useMemo(
    () => ({
      AttackVector: "AV",
      AttackComplexity: "AC",
      AttackRequirements: "AT",
      PrivilegesRequired: "PR",
      UserInteraction: "UI",
      Confidentiality: "VC",
      Integrity: "VI",
      Availability: "VA",
      SubsequentConfidentiality: "SC",
      SubsequentIntegrity: "SI",
      SubsequentAvailability: "SA",
      ExploitMaturity: "E",
      ConfidentialityRequirements: "CR",
      IntegrityRequirements: "IR",
      AvailabilityRequirements: "AR",
      ModifiedAttackVector: "MAV",
      ModifiedAttackComplexity: "MAC",
      ModifiedAttackRequirements: "MAT",
      ModifiedPrivilegesRequired: "MPR",
      ModifiedUserInteraction: "MUI",
      ModifiedConfidentiality: "MVC",
      ModifiedIntegrity: "MVI",
      ModifiedAvailability: "MVA",
      ModifiedSubsequentConfidentiality: "MSC",
      ModifiedSubsequentIntegrity: "MSI",
      ModifiedSubsequentAvailability: "MSA",
      Safety: "S",
      Automatable: "AU",
      Recovery: "R",
      ValueDensity: "V",
      ResponseEffort: "RE",
      ProviderUrgency: "U",
    }),
    []
  );
  const [cvssString, setCvssString] = useState<string>(value);
  const [cvss4Score, setCvss4Score] = useState(0);
  const [error, setError] = useState("");

  const validateCvssVector = (vector: string): boolean => {
    const vectorInstance = new Vector({ setError });
    const isValid = vectorInstance.validateStringVector(vector);

    return isValid;
  };

  const calculateRaw = metricsObj => {
    const baseString = "CVSS:4.0";
    const metricEntries = Object.entries(metricsObj)
      .filter(([, value]) => value !== "X")
      .map(([key, value]) => `/${metricLabelsShort[key]}:${value}`)
      .join("");
    return baseString + metricEntries;
  };

  useEffect(() => {
    if (!value) {
      return;
    }

    const prefixed = value.startsWith("CVSS:4.0/") ? value : "CVSS:4.0/" + value;
    if (!validateCvssVector(prefixed)) {
      setError("Invalid vector");
      return;
    }

    const vector = new Vector({ vectorString: prefixed });
    const parsedValues: Metrics = {
      AttackVector: vector.metrics.AV,
      AttackComplexity: vector.metrics.AC,
      AttackRequirements: vector.metrics.AT,
      PrivilegesRequired: vector.metrics.PR,
      UserInteraction: vector.metrics.UI,
      Confidentiality: vector.metrics.VC,
      Integrity: vector.metrics.VI,
      Availability: vector.metrics.VA,
      SubsequentConfidentiality: vector.metrics.SC,
      SubsequentIntegrity: vector.metrics.SI,
      SubsequentAvailability: vector.metrics.SA,
      ExploitMaturity: vector.metrics.E,
      ConfidentialityRequirements: vector.metrics.CR,
      IntegrityRequirements: vector.metrics.IR,
      AvailabilityRequirements: vector.metrics.AR,
      ModifiedAttackVector: vector.metrics.MAV,
      ModifiedAttackComplexity: vector.metrics.MAC,
      ModifiedAttackRequirements: vector.metrics.MAT,
      ModifiedPrivilegesRequired: vector.metrics.MPR,
      ModifiedUserInteraction: vector.metrics.MUI,
      ModifiedConfidentiality: vector.metrics.MVC,
      ModifiedIntegrity: vector.metrics.MVI,
      ModifiedAvailability: vector.metrics.MVA,
      ModifiedSubsequentConfidentiality: vector.metrics.MSC,
      ModifiedSubsequentIntegrity: vector.metrics.MSI,
      ModifiedSubsequentAvailability: vector.metrics.MSA,
      Safety: vector.metrics.S,
      Automatable: vector.metrics.AU,
      Recovery: vector.metrics.R,
      ValueDensity: vector.metrics.V,
      ResponseEffort: vector.metrics.RE,
      ProviderUrgency: vector.metrics.U,
    };

    setCvssString(prefixed);
    setMetrics(parsedValues);

    const instance = new CVSS40(prefixed);
    setCvss4Score(instance.calculateScore());

    setError("");
  }, [value]);

  useEffect(() => {
    const vectorString = calculateRaw(metrics);
    const instance = new CVSS40(vectorString);
    setCvss4Score(instance.calculateScore());
    setCvssString(vectorString);
    setError("");
  }, [metrics]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newCvssValue = e.target.value;
    const prefixedCvss = newCvssValue.startsWith("CVSS:4.0/") ? newCvssValue : "CVSS:4.0/" + newCvssValue;
    setCvssString(prefixedCvss);
    onChange?.(prefixedCvss);
    if (!validateCvssVector(prefixedCvss)) {
      return;
    }

    const vector = new Vector({ vectorString: prefixedCvss });
    const parsedValues: Metrics = {
      AttackVector: vector.metrics.AV,
      AttackComplexity: vector.metrics.AC,
      AttackRequirements: vector.metrics.AT,
      PrivilegesRequired: vector.metrics.PR,
      UserInteraction: vector.metrics.UI,
      Confidentiality: vector.metrics.VC,
      Integrity: vector.metrics.VI,
      Availability: vector.metrics.VA,
      SubsequentConfidentiality: vector.metrics.SC,
      SubsequentIntegrity: vector.metrics.SI,
      SubsequentAvailability: vector.metrics.SA,
      ExploitMaturity: vector.metrics.E,
      ConfidentialityRequirements: vector.metrics.CR,
      IntegrityRequirements: vector.metrics.IR,
      AvailabilityRequirements: vector.metrics.AR,
      ModifiedAttackVector: vector.metrics.MAV,
      ModifiedAttackComplexity: vector.metrics.MAC,
      ModifiedAttackRequirements: vector.metrics.MAT,
      ModifiedPrivilegesRequired: vector.metrics.MPR,
      ModifiedUserInteraction: vector.metrics.MUI,
      ModifiedConfidentiality: vector.metrics.MVC,
      ModifiedIntegrity: vector.metrics.MVI,
      ModifiedAvailability: vector.metrics.MVA,
      ModifiedSubsequentConfidentiality: vector.metrics.MSC,
      ModifiedSubsequentIntegrity: vector.metrics.MSI,
      ModifiedSubsequentAvailability: vector.metrics.MSA,
      Safety: vector.metrics.S,
      Automatable: vector.metrics.AU,
      Recovery: vector.metrics.R,
      ValueDensity: vector.metrics.V,
      ResponseEffort: vector.metrics.RE,
      ProviderUrgency: vector.metrics.U,
    };
    setMetrics(parsedValues);

    setError("");
  };

  const handleButtonClick = (key: string, value: string) => {
    setMetrics(prev => {
      const updatedMetrics = { ...prev, [key]: value };
      setCvssString(calculateRaw(updatedMetrics));
      onChange?.(calculateRaw(updatedMetrics));
      return updatedMetrics;
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
          label="CVSSv4.0 vector"
          id="cvssv4"
          value={cvssString}
          onChange={handleInputChange}
        />
        <ScoreBar score={cvss4Score} />
      </Grid>
      <Accordion title={"CVSSv4.0 Calculator"} getIsOpen={setIsAccordionOpen}>
        <CVSS40Render
          {...{
            selectedValues: metrics,
            handleButtonClick,
          }}
        />
      </Accordion>
    </div>
  );
}
