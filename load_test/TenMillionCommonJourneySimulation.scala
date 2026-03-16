package roving

import io.gatling.core.Predef._
import io.gatling.http.Predef._
import scala.util.Random
import scala.concurrent.duration._

class TenMillionCommonJourneySimulation extends Simulation {

  val referrersList: Seq[String] = Seq(
    "https://www.google.com/",
    "https://www.google.co.uk/",
    "https://www.google.ca/",
    "https://twitter.com/",
    "https://www.youtube.com/",
    "https://github.com/",
    "https://www.reddit.com/",
    "https://www.facebook.com/",
    "https://www.linkedin.com/",
    "https://www.instagram.com/",
    "https://susudong.com",
    "https://some-random-website.com"
  )

  val fireFox = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:92.0) Gecko/20100101 Firefox/92.0"
  val safari = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.0 Safari/605.1.15"
  val edge = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36 Edg/93.0.961.38"
  val internet11 = "Mozilla/5.0 (Windows NT 10.0; WOW64; Trident/7.0; AS; rv:11.0) like Gecko"
  val opera = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36 OPR/79.0.4143.50"
  val qqBroswerMini = "(MQQBrowser/Mini)(?:(\\d+)(?:\\.(\\d+)|)(?:\\.(\\d+)|)|)"
  val tentaBroswer = "(Tenta/)(\\d+)\\.(\\d+)\\.(\\d+)"

  def getRandomReferrer(referrers: Seq[String]): String = {
    referrers(Random.nextInt(referrers.length))
  }

  def randomUserAgent(): String = {
    val userAgents = Seq(
      fireFox,
      safari,
      edge,
      internet11,
      opera,
      qqBroswerMini,
      tentaBroswer
    )
    userAgents(Random.nextInt(userAgents.length))
  }

  // The base URL and the common headers
  val httpProtocol = http
    .baseUrl("http://localhost:80")
    .acceptHeader("application/json")
    //.userAgentHeader("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36")

  // Common paths for user journey
  val commonPaths = Seq("/sign-up", "/sign-in", "/settings", "/dashboard")
  val totalPaths = Seq("/sign-up", "/sign-in", "/settings", "/dashboard", "/home", "/activation", "/blog", "/press", "/profile", "/privacy-policy")

  def randomPath(): String = {
    totalPaths(Random.nextInt(totalPaths.length))
  }

  def generateSalt(): String = {
    Random.alphanumeric.take(13).mkString
  }

  val commonScenario = scenario("Common User Journey")
    .exec(session => {
      val newSession = session.set("ip", s"${Random.nextInt(255)}.${Random.nextInt(255)}.${Random.nextInt(255)}.${Random.nextInt(255)}")
      val updatedSessionWithReferrer = newSession.set("referrer", getRandomReferrer(referrersList))
      val updatedSessionWithUserAgent = updatedSessionWithReferrer.set("userAgent", randomUserAgent())
      val updatedSessionWithSalt = updatedSessionWithUserAgent.set("salt", generateSalt())
      updatedSessionWithSalt
    })
    .foreach(commonPaths, "path") {
      exec(session => {
        val updatedSession = session.set("timestamp", System.currentTimeMillis())
        updatedSession
      })
      .exec(
        http("Common Journey")
          .post("/api/event")
          .header("X-Forwarded-For", "#{ip}")
          .header("User-Agent", "#{userAgent}")
          .body(StringBody(
            """{
               |"e": "pageview",
               |"u": "https://justinsdong.dev#{path}",
               |"r": "#{referrer}",
               |"t": #{timestamp},
               |"s": "#{salt}"
               |}""".stripMargin))
      ).pause(1 second)
    }
  
  val randomScenario = scenario("Random User Journey")
    .exec(session => {
      val sessionWithIp = session.set("ip", s"${Random.nextInt(255)}.${Random.nextInt(255)}.${Random.nextInt(255)}.${Random.nextInt(255)}")
      val updatedSessionWithReferrer = sessionWithIp.set("referrer", getRandomReferrer(referrersList))
      val updatedSessionWithUserAgent = updatedSessionWithReferrer.set("userAgent", randomUserAgent())
      val updatedSessionWithSalt = updatedSessionWithUserAgent.set("salt", generateSalt())
      updatedSessionWithSalt
    })
    .repeat(4) {
      exec(session => {
        val sessionWithTimestamp = session.set("timestamp", System.currentTimeMillis())
        val sessionWithPath = sessionWithTimestamp.set("path", randomPath())
        sessionWithPath
      })
      .exec(http("Random Journey")
        .post("/api/event")
        .header("X-Forwarded-For", "#{ip}")
        .header("User-Agent", "#{userAgent}")
        .body(StringBody(
          """{
             |"e": "pageview",
             |"u": "https://justinsdong.dev#{path}",
             |"r": "#{referrer}",
             |"t": #{timestamp},
             |"s": "#{salt}"
             |}""".stripMargin)).asJson
      ).pause(1 second)
    }
  
  setUp(
    commonScenario.inject(rampUsers(150000) during (10 minute)),
    randomScenario.inject(rampUsers(100000) during (10 minute))
  ).protocols(httpProtocol)
}
