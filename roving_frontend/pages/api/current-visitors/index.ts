import type { NextApiRequest, NextApiResponse } from 'next';
import { BACKEND_URL } from '@/lib/api';

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  res.setHeader('Cache-Control', 'no-store');
  switch (req.method) {
    case 'GET':
      try {
        const { siteId } = req.query;

        if (!siteId) {
          return res
            .status(400)
            .json({ message: 'Required query parameters are missing.' });
        }

        const response = await fetch(
          `${BACKEND_URL}/api/current-visitors?siteId=${siteId}`
        );

        if (!response.ok) {
          console.error('Server responded with status:', response.status);
          return res
            .status(response.status)
            .json({ message: 'Error fetching data from server.' });
        }

        try {
          const data = await response.json();
          return res.status(200).json(data);
        } catch (error) {
          // This will catch any errors parsing the response as JSON.
          console.error('Error parsing response as JSON:', error);
          return res.status(500).json({ message: 'Internal Server Error' });
        }
      } catch (error) {
        console.error('Error fetching data:', error);
        return res.status(500).json({ message: 'Internal Server Error' });
      }

    // Handle other methods like POST, PUT, DELETE, etc., as needed.
    case 'POST':
      // Handle POST requests here.
      break;

    default:
      return res.status(405).json({ message: 'Method Not Allowed' });
  }
}
