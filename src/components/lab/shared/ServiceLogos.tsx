// Service logos component for displaying service information with logos

import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { LabSession } from "@/types/lab";

interface ServiceLogosProps {
  lab: LabSession;
}

// Default color scheme for services
const getServiceColor = (serviceId: string): string => {
  if (serviceId.includes('failing')) {
    return "bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200";
  }
  if (serviceId.includes('palette')) {
    return "bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200";
  }
  if (serviceId.includes('proxmox')) {
    return "bg-purple-100 text-purple-800 dark:bg-purple-900 dark:text-purple-200";
  }
  if (serviceId.includes('terraform')) {
    return "bg-orange-100 text-orange-800 dark:bg-orange-900 dark:text-orange-200";
  }
  if (serviceId.includes('guacamole')) {
    return "bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200";
  }
  return "bg-gray-100 text-gray-800 dark:bg-gray-900 dark:text-gray-200";
};

export function ServiceLogos({ lab }: ServiceLogosProps) {
  if (!lab.usedServices || lab.usedServices.length === 0) {
    return null;
  }

  return (
    <Card className="rounded-2xl">
      <CardHeader>
        <CardTitle className="text-base">Services Used</CardTitle>
      </CardHeader>
      <CardContent>
        <div className="flex flex-wrap gap-3">
          {lab.usedServices.map((service) => (
            <Badge 
              key={service.service_id} 
              variant="secondary" 
              className={`gap-2 ${getServiceColor(service.service_id)}`}
            >
              {service.logo ? (
                <img 
                  src={service.logo} 
                  alt={service.name}
                  className="w-4 h-4 object-contain"
                />
              ) : (
                <div className="w-4 h-4 bg-muted rounded" />
              )}
              {service.name}
            </Badge>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
