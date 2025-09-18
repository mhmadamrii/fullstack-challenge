import * as amqp from 'amqplib';
import { Provider } from '@nestjs/common';

export const RabbitMQProvider: Provider = {
  provide: 'RABBITMQ_CHANNEL',
  useFactory: async () => {
    const url = process.env.RABBITMQ_URL || 'amqp://guest:guest@localhost:5672';

    let conn: amqp.Connection | null = null;
    let channel: amqp.Channel | null = null;

    for (let i = 0; i < 5; i++) {
      // try up to 5 times
      try {
        conn = await amqp.connect(url);
        channel = await conn.createChannel();

        await channel.assertExchange('events', 'topic', { durable: true });
        await channel.assertQueue('order.created', { durable: true });
        await channel.assertQueue('product.created', { durable: true });
        await channel.bindQueue('product.created', 'events', 'product.created');
        await channel.bindQueue('order.created', 'events', 'order.created');

        console.log('✅ Connected to RabbitMQ');
        break;
      } catch (err) {
        console.error(`❌ RabbitMQ not ready, retrying in 5s... (${i + 1}/5)`);
        await new Promise((res) => setTimeout(res, 5000));
      }
    }

    if (!channel) {
      throw new Error('Could not connect to RabbitMQ after 5 attempts');
    }

    return channel;
  },
};
