import * as amqp from 'amqplib';
import { Provider } from '@nestjs/common';

export const RabbitMQProvider: Provider = {
  provide: 'RABBITMQ_CHANNEL',
  useFactory: async () => {
    const url = 'amqp://guest:guest@localhost:5672';
    const conn = await amqp.connect(url);
    const channel = await conn.createChannel();

    await channel.assertExchange('events', 'topic', { durable: true });
    await channel.assertQueue('order.created', { durable: true });
    await channel.assertQueue('product.created', { durable: true });
    await channel.assertQueue('something.created', { durable: true });
    await channel.bindQueue('product.created', 'events', 'product.created');
    await channel.bindQueue('order.created', 'events', 'order.created');

    return channel;
  },
};
