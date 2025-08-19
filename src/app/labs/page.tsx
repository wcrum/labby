'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';

import { Clock, Server, Users, Zap } from 'lucide-react';
import { apiService, LabTemplate } from '@/lib/api';
import { AppLayout } from '@/components/layout/AppLayout';

export default function LabsPage() {
  const [templates, setTemplates] = useState<LabTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [creatingLab, setCreatingLab] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    loadTemplates();
  }, []);

  const loadTemplates = async () => {
    try {
      setLoading(true);
      const data = await apiService.getTemplates();
      setTemplates(data);
    } catch (err) {
      setError('Failed to load lab templates');
      console.error('Error loading templates:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateLab = async (templateId: string) => {
    try {
      setCreatingLab(templateId);
      const lab = await apiService.createLabFromTemplate(templateId);
      router.push(`/lab?id=${lab.id}`);
    } catch (err) {
      setError('Failed to create lab');
      console.error('Error creating lab:', err);
    } finally {
      setCreatingLab(null);
    }
  };

  const formatDuration = (duration: string) => {
    // Convert duration like "72h" to "3 days"
    const hours = parseInt(duration.replace('h', ''));
    if (hours >= 24) {
      const days = Math.floor(hours / 24);
      return `${days} day${days > 1 ? 's' : ''}`;
    }
    return `${hours} hour${hours > 1 ? 's' : ''}`;
  };

  const getServiceIcon = (type: string) => {
    switch (type) {
      case 'palette':
        return <Server className="h-4 w-4" />;
      case 'proxmox':
        return <Zap className="h-4 w-4" />;
      default:
        return <Server className="h-4 w-4" />;
    }
  };

  if (loading) {
    return (
      <AppLayout>
        <div className="container mx-auto px-6 py-6">
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">Loading lab templates...</p>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  if (error) {
    return (
      <AppLayout>
        <div className="container mx-auto px-6 py-6">
          <div className="flex items-center justify-center h-64">
            <div className="text-center">
              <p className="text-destructive mb-4">{error}</p>
              <Button onClick={loadTemplates} variant="outline">
                Try Again
              </Button>
            </div>
          </div>
        </div>
      </AppLayout>
    );
  }

  return (
    <AppLayout>
      <div className="container mx-auto px-6 py-6">
        <div className="max-w-6xl mx-auto space-y-6">
        {/* Header */}
        <div className="text-center space-y-2">
          <h1 className="text-3xl font-bold">Available Labs</h1>
          <p className="text-muted-foreground">
            Choose a lab template to get started with your learning environment
          </p>
        </div>

        {/* Templates Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {templates.map((template) => (
            <Card key={template.id} className="hover:shadow-lg transition-shadow">
              <CardHeader>
                <div className="flex items-start justify-between">
                  <div className="space-y-1">
                    <CardTitle className="text-lg">{template.name}</CardTitle>
                    <CardDescription>{template.description}</CardDescription>
                  </div>
                  <Badge variant="outline" className="text-xs">
                    {formatDuration(template.expiration_duration)}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                {/* Services */}
                <div className="space-y-2">
                  <h4 className="text-sm font-medium text-muted-foreground">Services included:</h4>
                  <div className="space-y-1">
                    {template.services.map((service, index) => (
                      <div key={index} className="flex items-center gap-2 text-sm">
                        {getServiceIcon(service.type)}
                        <span>{service.name}</span>
                        <Badge variant="secondary" className="text-xs">
                          {service.type}
                        </Badge>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Template Info */}
                <div className="flex items-center justify-between text-xs text-muted-foreground">
                  <div className="flex items-center gap-1">
                    <Clock className="h-3 w-3" />
                    <span>Created {new Date(template.created_at).toLocaleDateString()}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Users className="h-3 w-3" />
                    <span>{template.owner}</span>
                  </div>
                </div>

                {/* Action Button */}
                <Button
                  onClick={() => handleCreateLab(template.id)}
                  disabled={creatingLab === template.id}
                  className="w-full"
                >
                  {creatingLab === template.id ? (
                    <>
                      <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                      Creating Lab...
                    </>
                  ) : (
                    'Start Lab'
                  )}
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>

        {templates.length === 0 && (
          <div className="text-center py-12">
            <Server className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h3 className="text-lg font-medium mb-2">No lab templates available</h3>
            <p className="text-muted-foreground">
              Contact your administrator to configure lab templates.
            </p>
          </div>
        )}
        </div>
      </div>
    </AppLayout>
  );
}
