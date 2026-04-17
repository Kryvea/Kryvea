const CVSSVersionIdentifier = "CVSS:3.1";
const exploitabilityCoefficient = 8.22;
const scopeCoefficient = 1.08;

const vectorStringRegex_31 =
  /^CVSS:3\.1\/((AV:[NALP]|AC:[LH]|PR:[UNLH]|UI:[NR]|S:[UC]|[CIA]:[NLH]|E:[XUPFH]|RL:[XOTWU]|RC:[XURC]|[CIA]R:[XLMH]|MAV:[XNALP]|MAC:[XLH]|MPR:[XUNLH]|MUI:[XNR]|MS:[XUC]|M[CIA]:[XNLH])\/)*(AV:[NALP]|AC:[LH]|PR:[UNLH]|UI:[NR]|S:[UC]|[CIA]:[NLH]|E:[XUPFH]|RL:[XOTWU]|RC:[XURC]|[CIA]R:[XLMH]|MAV:[XNALP]|MAC:[XLH]|MPR:[XUNLH]|MUI:[XNR]|MS:[XUC]|M[CIA]:[XNLH])$/;

const Weight: Record<string, any> = {
  AV: { N: 0.85, A: 0.62, L: 0.55, P: 0.2 },
  AC: { H: 0.44, L: 0.77 },
  PR: {
    U: { N: 0.85, L: 0.62, H: 0.27 },
    C: { N: 0.85, L: 0.68, H: 0.5 },
  },
  UI: { N: 0.85, R: 0.62 },
  S: { U: 6.42, C: 7.52 },
  CIA: { N: 0, L: 0.22, H: 0.56 },
  E: { X: 1, U: 0.91, P: 0.94, F: 0.97, H: 1 },
  RL: { X: 1, O: 0.95, T: 0.96, W: 0.97, U: 1 },
  RC: { X: 1, U: 0.92, R: 0.96, C: 1 },
  CIAR: { X: 1, L: 0.5, M: 1, H: 1.5 },
};

const severityRatings = [
  { name: "None", bottom: 0.0, top: 0.0 },
  { name: "Low", bottom: 0.1, top: 3.9 },
  { name: "Medium", bottom: 4.0, top: 6.9 },
  { name: "High", bottom: 7.0, top: 8.9 },
  { name: "Critical", bottom: 9.0, top: 10.0 },
];

function roundUp1(input: number): number {
  const intInput = Math.round(input * 100000);
  return intInput % 10000 === 0 ? intInput / 100000 : Math.ceil(intInput / 10000) / 10;
}

function severityRating(score: number): string {
  for (let i = 0; i < severityRatings.length; i++) {
    if (score >= severityRatings[i].bottom && score <= severityRatings[i].top) {
      return severityRatings[i].name;
    }
  }
  return "Unknown";
}

type Metrics = {
  AttackVector: string;
  AttackComplexity: string;
  PrivilegesRequired: string;
  UserInteraction: string;
  Scope: string;
  Confidentiality: string;
  Integrity: string;
  Availability: string;
  ExploitCodeMaturity: string;
  RemediationLevel: string;
  ReportConfidence: string;
  ConfidentialityRequirement: string;
  IntegrityRequirement: string;
  AvailabilityRequirement: string;
  ModifiedAttackVector: string;
  ModifiedAttackComplexity: string;
  ModifiedPrivilegesRequired: string;
  ModifiedUserInteraction: string;
  ModifiedScope: string;
  ModifiedConfidentiality: string;
  ModifiedIntegrity: string;
  ModifiedAvailability: string;
};
type CvssCalcFail = { success: false; errorType: string; errorMetrics?: string[] };
type CvssCalcSuccess = {
  success: true;
  metrics: Metrics;
  baseMetricScore: number;
  baseSeverity: string;
  baseISS: number;
  baseImpact: number;
  baseExploitability: number;
  temporalMetricScore: number;
  temporalSeverity: string;
  environmentalMetricScore: number;
  environmentalSeverity: string;
  environmentalMISS: number;
  environmentalModifiedImpact: number;
  environmentalModifiedExploitability: number;
  vectorString: string;
};
export function calculateCVSSFromMetrics(metrics): CvssCalcSuccess | CvssCalcFail {
  const {
    AttackVector,
    AttackComplexity,
    PrivilegesRequired,
    UserInteraction,
    Scope,
    Confidentiality,
    Integrity,
    Availability,
    ExploitCodeMaturity,
    RemediationLevel,
    ReportConfidence,
    ConfidentialityRequirement,
    IntegrityRequirement,
    AvailabilityRequirement,
    ModifiedAttackVector,
    ModifiedAttackComplexity,
    ModifiedPrivilegesRequired,
    ModifiedUserInteraction,
    ModifiedScope,
    ModifiedConfidentiality,
    ModifiedIntegrity,
    ModifiedAvailability,
  } = metrics;

  const badMetrics: string[] = [];

  if (!AttackVector) badMetrics.push("AV");
  if (!AttackComplexity) badMetrics.push("AC");
  if (!PrivilegesRequired) badMetrics.push("PR");
  if (!UserInteraction) badMetrics.push("UI");
  if (!Scope) badMetrics.push("S");
  if (!Confidentiality) badMetrics.push("C");
  if (!Integrity) badMetrics.push("I");
  if (!Availability) badMetrics.push("A");

  if (badMetrics.length > 0) {
    return { success: false, errorType: "Missing Base Metric", errorMetrics: badMetrics };
  }

  const AV = AttackVector;
  const AC = AttackComplexity;
  const PR = PrivilegesRequired;
  const UI = UserInteraction;
  const S = Scope;
  const C = Confidentiality;
  const I = Integrity;
  const A = Availability;

  const E = ExploitCodeMaturity || "X";
  const RL = RemediationLevel || "X";
  const RC = ReportConfidence || "X";

  const CR = ConfidentialityRequirement || "X";
  const IR = IntegrityRequirement || "X";
  const AR = AvailabilityRequirement || "X";
  const MAV = ModifiedAttackVector || "X";
  const MAC = ModifiedAttackComplexity || "X";
  const MPR = ModifiedPrivilegesRequired || "X";
  const MUI = ModifiedUserInteraction || "X";
  const MS = ModifiedScope || "X";
  const MC = ModifiedConfidentiality || "X";
  const MI = ModifiedIntegrity || "X";
  const MA = ModifiedAvailability || "X";

  if (!Weight.AV.hasOwnProperty(AV)) badMetrics.push("AV");
  if (!Weight.AC.hasOwnProperty(AC)) badMetrics.push("AC");
  if (!Weight.PR.U.hasOwnProperty(PR)) badMetrics.push("PR");
  if (!Weight.UI.hasOwnProperty(UI)) badMetrics.push("UI");
  if (!Weight.S.hasOwnProperty(S)) badMetrics.push("S");
  if (!Weight.CIA.hasOwnProperty(C)) badMetrics.push("C");
  if (!Weight.CIA.hasOwnProperty(I)) badMetrics.push("I");
  if (!Weight.CIA.hasOwnProperty(A)) badMetrics.push("A");
  if (!Weight.E.hasOwnProperty(E)) badMetrics.push("E");
  if (!Weight.RL.hasOwnProperty(RL)) badMetrics.push("RL");
  if (!Weight.RC.hasOwnProperty(RC)) badMetrics.push("RC");

  if (CR !== "X" && !Weight.CIAR.hasOwnProperty(CR)) badMetrics.push("CR");
  if (IR !== "X" && !Weight.CIAR.hasOwnProperty(IR)) badMetrics.push("IR");
  if (AR !== "X" && !Weight.CIAR.hasOwnProperty(AR)) badMetrics.push("AR");
  if (MAV !== "X" && !Weight.AV.hasOwnProperty(MAV)) badMetrics.push("MAV");
  if (MAC !== "X" && !Weight.AC.hasOwnProperty(MAC)) badMetrics.push("MAC");
  if (MPR !== "X" && !Weight.PR.U.hasOwnProperty(MPR)) badMetrics.push("MPR");
  if (MUI !== "X" && !Weight.UI.hasOwnProperty(MUI)) badMetrics.push("MUI");
  if (MS !== "X" && !Weight.S.hasOwnProperty(MS)) badMetrics.push("MS");
  if (MC !== "X" && !Weight.CIA.hasOwnProperty(MC)) badMetrics.push("MC");
  if (MI !== "X" && !Weight.CIA.hasOwnProperty(MI)) badMetrics.push("MI");
  if (MA !== "X" && !Weight.CIA.hasOwnProperty(MA)) badMetrics.push("MA");

  if (badMetrics.length > 0) {
    return { success: false, errorType: "Unknown Metric Value", errorMetrics: badMetrics };
  }

  const metricWeightAV = Weight.AV[AV];
  const metricWeightAC = Weight.AC[AC];
  const metricWeightPR = Weight.PR[S][PR];
  const metricWeightUI = Weight.UI[UI];
  const metricWeightS = Weight.S[S];
  const metricWeightC = Weight.CIA[C];
  const metricWeightI = Weight.CIA[I];
  const metricWeightA = Weight.CIA[A];
  const metricWeightE = Weight.E[E];
  const metricWeightRL = Weight.RL[RL];
  const metricWeightRC = Weight.RC[RC];
  const metricWeightCR = Weight.CIAR[CR];
  const metricWeightIR = Weight.CIAR[IR];
  const metricWeightAR = Weight.CIAR[AR];
  const metricWeightMAV = Weight.AV[MAV !== "X" ? MAV : AV];
  const metricWeightMAC = Weight.AC[MAC !== "X" ? MAC : AC];
  const metricWeightMPR = Weight.PR[MS !== "X" ? MS : S][MPR !== "X" ? MPR : PR];
  const metricWeightMUI = Weight.UI[MUI !== "X" ? MUI : UI];
  const metricWeightMS = Weight.S[MS !== "X" ? MS : S];
  const metricWeightMC = Weight.CIA[MC !== "X" ? MC : C];
  const metricWeightMI = Weight.CIA[MI !== "X" ? MI : I];
  const metricWeightMA = Weight.CIA[MA !== "X" ? MA : A];

  let iss: number;
  let impact: number;
  let exploitability: number;
  let baseScore: number;

  iss = 1 - (1 - metricWeightC) * (1 - metricWeightI) * (1 - metricWeightA);
  if (S === "U") {
    impact = metricWeightS * iss;
  } else {
    impact = metricWeightS * (iss - 0.029) - 3.25 * Math.pow(iss - 0.02, 15);
  }

  exploitability = exploitabilityCoefficient * metricWeightAV * metricWeightAC * metricWeightPR * metricWeightUI;

  if (impact <= 0) {
    baseScore = 0;
  } else {
    if (S === "U") {
      baseScore = roundUp1(Math.min(exploitability + impact, 10));
    } else {
      baseScore = roundUp1(Math.min(scopeCoefficient * (exploitability + impact), 10));
    }
  }

  const temporalScore = Math.min(baseScore * Weight.E[E] * Weight.RL[RL] * Weight.RC[RC], 10);

  let miss: number;
  let modifiedImpact: number;
  let envScore: number;
  let modifiedExploitability: number;

  miss = Math.min(
    1 -
      (1 - metricWeightMC * metricWeightCR) *
        (1 - metricWeightMI * metricWeightIR) *
        (1 - metricWeightMA * metricWeightAR),
    0.915
  );

  if (MS === "U" || (MS === "X" && S === "U")) {
    modifiedImpact = metricWeightMS * miss;
  } else {
    modifiedImpact = metricWeightMS * (miss - 0.029) - 3.25 * Math.pow(miss * 0.9731 - 0.02, 13);
  }

  modifiedExploitability =
    exploitabilityCoefficient * metricWeightMAV * metricWeightMAC * metricWeightMPR * metricWeightMUI;

  if (modifiedImpact <= 0) {
    envScore = 0;
  } else if (MS === "U" || (MS === "X" && S === "U")) {
    envScore = roundUp1(
      roundUp1(Math.min(modifiedImpact + modifiedExploitability, 10)) * metricWeightE * metricWeightRL * metricWeightRC
    );
  } else {
    envScore = roundUp1(
      roundUp1(Math.min(scopeCoefficient * (modifiedImpact + modifiedExploitability), 10)) *
        metricWeightE *
        metricWeightRL *
        metricWeightRC
    );
  }

  let vectorString =
    CVSSVersionIdentifier +
    "/AV:" +
    AV +
    "/AC:" +
    AC +
    "/PR:" +
    PR +
    "/UI:" +
    UI +
    "/S:" +
    S +
    "/C:" +
    C +
    "/I:" +
    I +
    "/A:" +
    A;

  if (E !== "X" || RL !== "X" || RC !== "X") {
    vectorString += "/E:" + E;
    vectorString += "/RL:" + RL;
    vectorString += "/RC:" + RC;
  }
  if (
    CR !== "X" ||
    IR !== "X" ||
    AR !== "X" ||
    MAV !== "X" ||
    MAC !== "X" ||
    MPR !== "X" ||
    MUI !== "X" ||
    MS !== "X" ||
    MC !== "X" ||
    MI !== "X" ||
    MA !== "X"
  ) {
    vectorString += "/CR:" + CR;
    vectorString += "/IR:" + IR;
    vectorString += "/AR:" + AR;
    vectorString += "/MAV:" + MAV;
    vectorString += "/MAC:" + MAC;
    vectorString += "/MPR:" + MPR;
    vectorString += "/MUI:" + MUI;
    vectorString += "/MS:" + MS;
    vectorString += "/MC:" + MC;
    vectorString += "/MI:" + MI;
    vectorString += "/MA:" + MA;
  }
  return {
    success: true,
    metrics,
    baseMetricScore: parseFloat(baseScore.toFixed(1)),
    baseSeverity: severityRating(baseScore),
    baseISS: iss,
    baseImpact: impact,
    baseExploitability: exploitability,
    temporalMetricScore: parseFloat(temporalScore.toFixed(1)),
    temporalSeverity: severityRating(temporalScore),
    environmentalMetricScore: parseFloat(envScore.toFixed(1)),
    environmentalSeverity: severityRating(envScore),
    environmentalMISS: miss,
    environmentalModifiedImpact: modifiedImpact,
    environmentalModifiedExploitability: modifiedExploitability,
    vectorString: vectorString,
  };
}

export function calculateCVSSFromVector(vectorString: string): CvssCalcFail | CvssCalcSuccess {
  const metricValues: Record<string, string | undefined> = {
    AV: undefined,
    AC: undefined,
    PR: undefined,
    UI: undefined,
    S: undefined,
    C: undefined,
    I: undefined,
    A: undefined,
    E: undefined,
    RL: undefined,
    RC: undefined,
    CR: undefined,
    IR: undefined,
    AR: undefined,
    MAV: undefined,
    MAC: undefined,
    MPR: undefined,
    MUI: undefined,
    MS: undefined,
    MC: undefined,
    MI: undefined,
    MA: undefined,
  };
  const badMetrics: string[] = [];
  if (!vectorStringRegex_31.test(vectorString)) {
    return { success: false, errorType: "Malformed Vector String" };
  }
  const metricNameValue = vectorString.substring(CVSSVersionIdentifier.length).split("/");
  for (const i in metricNameValue) {
    if (metricNameValue.hasOwnProperty(i)) {
      const singleMetric = metricNameValue[i].split(":");
      if (typeof metricValues[singleMetric[0]] === "undefined") {
        metricValues[singleMetric[0]] = singleMetric[1];
      } else {
        badMetrics.push(singleMetric[0]);
      }
    }
  }

  for (const key in metricValues) {
    if (metricValues[key] === undefined) {
      metricValues[key] = "X";
    }
  }

  if (badMetrics.length > 0) {
    return { success: false, errorType: "Multiple Definitions Of Metrics", errorMetrics: badMetrics };
  }

  const metrics = {
    AttackVector: metricValues.AV,
    AttackComplexity: metricValues.AC,
    PrivilegesRequired: metricValues.PR,
    UserInteraction: metricValues.UI,
    Scope: metricValues.S,
    Confidentiality: metricValues.C,
    Integrity: metricValues.I,
    Availability: metricValues.A,
    ExploitCodeMaturity: metricValues.E,
    RemediationLevel: metricValues.RL,
    ReportConfidence: metricValues.RC,
    ConfidentialityRequirement: metricValues.CR,
    IntegrityRequirement: metricValues.IR,
    AvailabilityRequirement: metricValues.AR,
    ModifiedAttackVector: metricValues.MAV,
    ModifiedAttackComplexity: metricValues.MAC,
    ModifiedPrivilegesRequired: metricValues.MPR,
    ModifiedUserInteraction: metricValues.MUI,
    ModifiedScope: metricValues.MS,
    ModifiedConfidentiality: metricValues.MC,
    ModifiedIntegrity: metricValues.MI,
    ModifiedAvailability: metricValues.MA,
  };

  return calculateCVSSFromMetrics(metrics);
}
