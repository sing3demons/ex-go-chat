import { LoadingSpinner } from './LoadingSpinner';

interface PageLoaderProps {
  text?: string;
}

export const PageLoader = ({ text = 'Loading...' }: PageLoaderProps) => {
  return (
    <div className="fixed inset-0 bg-white bg-opacity-90 flex items-center justify-center z-50">
      <div className="bg-white p-8 rounded-lg shadow-lg">
        <LoadingSpinner size="lg" color="blue" text={text} />
      </div>
    </div>
  );
};
