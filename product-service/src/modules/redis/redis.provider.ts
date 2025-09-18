import { createClient } from 'redis';
import { Provider } from '@nestjs/common';

export const RedisProvider: Provider = {
  provide: 'REDIS_CLIENT',
  useFactory: async () => {
    const client = createClient({
      url: process.env.REDIS_URL || 'redis://localhost:6379',
    });

    client.on('error', (err) => console.error('Redis Client Error', err));

    await client.connect();
    console.log('Redis connected successfully');
    return client;
  },
};
