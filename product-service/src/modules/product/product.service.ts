import { Injectable, Inject, NotFoundException } from '@nestjs/common';
import { CreateProductDto } from './product.dto';
import { PrismaService } from '../../prisma.service';
import type { RedisClientType } from 'redis';
import * as amqp from 'amqplib';

@Injectable()
export class ProductService {
  constructor(
    private prisma: PrismaService,
    @Inject('REDIS_CLIENT') private redis: RedisClientType,
    @Inject('RABBITMQ_CHANNEL') private readonly channel: amqp.Channel,
  ) {
    this.listenToOrderCreated();
  }

  async listenToOrderCreated() {
    this.channel.consume('order.created', async (msg) => {
      if (!msg) return;

      try {
        const order = JSON.parse(msg.content.toString());
        console.log('📦 Order created:', order);

        const updatedProduct = await this.prisma.products.updateMany({
          where: { id: order.productId, qty: { gt: 0 } },
          data: { qty: { decrement: 1 } },
        });

        if (updatedProduct.count === 0) {
          console.log(
            `⚠️ Product ${order.productId} is out of stock or update failed`,
          );
        } else {
          console.log(`✅ Product ${order.productId} stock decremented`);

          await this.redis.del(`product_${order.productId}`);
          await this.redis.del('products');
        }

        this.channel.ack(msg);
      } catch (err) {
        console.error('❌ Failed to process order.created:', err);
        this.channel.ack(msg); // ack anyway to prevent infinite retries
      }
    });
  }

  async createProduct(dto: CreateProductDto) {
    const newProduct = await this.prisma.products.create({ data: dto });
    await this.redis.del('products');

    const payload = Buffer.from(JSON.stringify(newProduct));
    await this.channel.publish('events', 'product.created', payload);
    return newProduct;
  }

  async getAllProducts() {
    const cachedProducts = await this.redis.get('products');
    if (cachedProducts) {
      console.log('Cached products found ✅');
      return JSON.parse(cachedProducts);
    }

    console.log('Cached products missing 🗑️');
    const products = await this.prisma.products.findMany();
    await this.redis.set('products', JSON.stringify(products));
    return products;
  }

  async getProductById(id: string) {
    const cachedProduct = await this.redis.get(`product_${id}`);
    if (cachedProduct) {
      console.log(`✅ Cached product found for id:${id}`);
      return JSON.parse(cachedProduct);
    }
    const product = await this.prisma.products.findUnique({ where: { id } });
    if (!product) {
      throw new NotFoundException(`Product with id ${id} not found`);
    }
    await this.redis.set(`product_${id}`, JSON.stringify(product));
    console.log(`🗑️ Cached product missed for id:${id}, set new cache`);
    return product;
  }

  async createProductWithoutCache(dto: CreateProductDto) {
    return this.prisma.products.create({ data: dto });
  }

  async getAllProductsWithoutCache() {
    return this.prisma.products.findMany();
  }
}
