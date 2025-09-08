import { cn } from "@/lib/utils";

interface LoadingSkeletonProps {
  className?: string;
}

export function LoadingSkeleton({ className }: LoadingSkeletonProps) {
  return (
    <div
      className={cn(
        "animate-pulse rounded-md bg-muted",
        className
      )}
    />
  );
}

// Predefined skeleton variants for common use cases
export function PageHeaderSkeleton() {
  return (
    <div className="space-y-2">
      <LoadingSkeleton className="h-8 w-64" />
      <LoadingSkeleton className="h-4 w-96" />
    </div>
  );
}

export function CardSkeleton() {
  return (
    <div className="rounded-xl border bg-card p-6">
      <div className="space-y-4">
        <div className="flex items-center space-x-4">
          <LoadingSkeleton className="h-12 w-12 rounded-full" />
          <div className="space-y-2">
            <LoadingSkeleton className="h-4 w-32" />
            <LoadingSkeleton className="h-3 w-24" />
          </div>
        </div>
        <LoadingSkeleton className="h-4 w-full" />
        <LoadingSkeleton className="h-4 w-3/4" />
      </div>
    </div>
  );
}

export function LabCardSkeleton() {
  return (
    <div className="rounded-xl border bg-card p-6">
      <div className="space-y-4">
        <div className="flex items-start justify-between">
          <div className="space-y-2">
            <LoadingSkeleton className="h-6 w-48" />
            <LoadingSkeleton className="h-4 w-32" />
          </div>
          <LoadingSkeleton className="h-6 w-16 rounded-full" />
        </div>
        <div className="space-y-2">
          <LoadingSkeleton className="h-4 w-full" />
          <LoadingSkeleton className="h-4 w-2/3" />
        </div>
        <div className="flex items-center space-x-4">
          <LoadingSkeleton className="h-8 w-20" />
          <LoadingSkeleton className="h-8 w-24" />
        </div>
      </div>
    </div>
  );
}

export function TableSkeleton({ rows = 5 }: { rows?: number }) {
  return (
    <div className="space-y-3">
      {/* Header */}
      <div className="flex space-x-4">
        <LoadingSkeleton className="h-4 w-32" />
        <LoadingSkeleton className="h-4 w-24" />
        <LoadingSkeleton className="h-4 w-20" />
        <LoadingSkeleton className="h-4 w-16" />
      </div>
      {/* Rows */}
      {Array.from({ length: rows }).map((_, i) => (
        <div key={i} className="flex space-x-4">
          <LoadingSkeleton className="h-4 w-32" />
          <LoadingSkeleton className="h-4 w-24" />
          <LoadingSkeleton className="h-4 w-20" />
          <LoadingSkeleton className="h-4 w-16" />
        </div>
      ))}
    </div>
  );
}

export function LabGridSkeleton() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {Array.from({ length: 4 }).map((_, i) => (
        <LabCardSkeleton key={i} />
      ))}
    </div>
  );
}

export function LabPageSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <PageHeaderSkeleton />
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
        {Array.from({ length: 2 }).map((_, i) => (
          <LoadingSkeleton key={i} className="h-56 rounded-2xl" />
        ))}
      </div>
    </div>
  );
}

export function AdminPageSkeleton() {
  return (
    <div className="p-6 space-y-6">
      <PageHeaderSkeleton />
      
      {/* Stats Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
      
      {/* Labs Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }).map((_, i) => (
          <LabCardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}

export function OrganizationsPageSkeleton() {
  return (
    <div className="p-6 space-y-6">
      <PageHeaderSkeleton />
      <div className="space-y-6">
        {Array.from({ length: 3 }).map((_, i) => (
          <CardSkeleton key={i} />
        ))}
      </div>
    </div>
  );
}

export function UsersPageSkeleton() {
  return (
    <div className="p-6 space-y-6">
      <PageHeaderSkeleton />
      <div className="rounded-xl border bg-card p-6">
        <TableSkeleton rows={8} />
      </div>
    </div>
  );
}
