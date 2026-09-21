import React from "react";
import { VerificationStatus as StatusType } from "../types";

interface Props {
  status: StatusType;
}

const statusColors: Record<StatusType, string> = {
  CREATED: "text-gray-500",
  COLLECTING: "text-blue-500",
  PROCESSING: "text-blue-500",
  EVALUATING: "text-purple-500",
  REVIEW_REQUIRED: "text-orange-500",
  REVIEWING: "text-orange-600",
  MORE_INFORMATION_REQUIRED: "text-yellow-600",
  APPROVED: "text-green-600",
  REJECTED: "text-red-600",
  FAILED: "text-red-700",
  CANCELLED: "text-gray-400",
  EXPIRED: "text-gray-400",
  COMPLETED: "text-green-500",
};

export const VerificationStatus: React.FC<Props> = ({ status }) => {
  return (
    <div className={`font-semibold flex items-center gap-2 ${statusColors[status] || "text-black"}`}>
      <span className="w-2 h-2 rounded-full bg-current inline-block"></span>
      {status}
    </div>
  );
};
