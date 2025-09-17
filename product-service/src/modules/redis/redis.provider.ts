import { createClient } from 'redis';
import { Provider } from '@nestjs/common';

export const RedisProvider: Provider = {
  provide: 'REDIS_CLIENT',
  useFactory: async () => {
    const client = createClient({
      username: 'default',
      password: process.env.REDIS_CLIENT_PASS,
      socket: {
        host: process.env.REDIS_HOST,
        port: process.env.REDIS_PORT as unknown as number,
      },
    });

    client.on('error', (err) => console.error('Redis Client Error', err));

    await client.connect();
    return client;
  },
};
